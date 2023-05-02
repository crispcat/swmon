package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileType int

const (
	HostsConfigFile FileType = iota
	HostsModelFile
)

func ParseConfig(path string) SwmonConfig {

	file, err := os.ReadFile(path)
	if err != nil {
		ErrorAll("Unable to open config file on path %s: %s", path, err)
	}

	var cfg SwmonConfig
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		ErrorAll("Unable to unmarshall config file on path %s: %s", path, err)
	}

	return cfg
}

func CreateConfigIfNExist(configPath string) {

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		strd, _ := yaml.Marshal(NEW_CONFIG)
		err = os.WriteFile(configPath, strd, os.FileMode(OS_FILE_PERMISSIONS_STRICT).Perm())
		if err != nil {
			WriteAll("Unable to create config file %s: %s", configPath, err)
		}
		WriteAll("Config created on path %s!", configPath)
		DoneLogs()
		os.Exit(0)
	}
}

func GetBackupFileName(ftype FileType) string {

	nowTime := time.Now().Format("2006_01_02__15_04_05")
	switch ftype {
	case HostsModelFile:
		return HOSTS_MODEL_FILE_BACKUP_PREFIX + nowTime + ".json"
	case HostsConfigFile:
		return HOSTS_CONFIG_FILE_BACKUP_PREFIX + nowTime + ".cfg"
	default:
		panic("Unknown file type!")
	}
}

func MakeBackupDir() error {

	if _, err := os.Stat(BACKUP_ROOT); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(BACKUP_ROOT, os.FileMode(OS_DIR_PERMISSIONS).Perm())
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteHostsModel(hostsMap *HostsModel) error {

	strd, err := json.MarshalIndent(hostsMap.Export(), "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	err = os.WriteFile(HOSTS_MODEL_FILE, strd, os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
	if err != nil {
		return err
	}

	err = MakeBackupDir()
	if err != nil {
		return err
	}

	err = os.WriteFile(GetBackupFileName(HostsConfigFile), strd, os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
	if err != nil {
		return err
	}

	return nil
}

func ReadHostsModel() (map[uint32]*Host, error) {

	file, err := os.ReadFile(HOSTS_MODEL_FILE)
	if err != nil {
		WriteAll("Unable to open hosts model file on path %s: %s", HOSTS_MODEL_FILE, err)
		return nil, err
	}

	var hostsModelMap map[uint32]*Host
	err = json.Unmarshal(file, &hostsModelMap)
	if err != nil {
		WriteAll("Unable to unmarshal hosts model file on path %s: %s", HOSTS_MODEL_FILE, err)
		return nil, err
	}

	return hostsModelMap, nil
}

func WriteNagiosHostsConfigFile(hostsModel *HostsModel) error {

	var hosts strings.Builder

	hosts.WriteString(NAGIOS_CONFIG_HEADER)

	hostsModel.Map.Range(func(key any, value any) bool {

		host := value.(*Host)
		if host.WriteToConfig {
			host := ConstructSwitchTemplate(host)
			hosts.WriteString(host)
		}

		return true
	})

	hostsModel.Conn.Range(func(key any, value any) bool {

		conn := value.(Connection)
		sg := ConstructConnectionServiceGroup(conn, hostsModel)
		hosts.WriteString(sg)

		return true
	})

	if hosts.Len() == 0 {
		WriteAll("No SNMP hosts found. Nothing will be written to %s.", HOSTS_CONFIG_FILE)
		return nil
	}

	hostsString := []byte(hosts.String())

	err := os.WriteFile(HOSTS_CONFIG_FILE, hostsString, os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
	if err != nil {
		return err
	}

	err = MakeBackupDir()
	if err != nil {
		return err
	}

	err = os.WriteFile(GetBackupFileName(HostsModelFile), hostsString, os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
	if err != nil {
		return err
	}

	return nil
}

func ReadMapFile(path string) []string {

	f, err := os.ReadFile(path)
	if err != nil {
		ErrorAll("MAPPER: UNABLE TO READ MAP FILE %s: %s", path, err.Error())
	}

	chunks := strings.Fields(string(f))
	return chunks
}

func WriteMapFile(path string, es []NagvisMapEntity, backup bool) {

	// backup original file
	if backup {
		backupDir := filepath.Dir(path) + "/maps_backup"
		if _, err := os.Stat(backupDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(backupDir, os.FileMode(OS_DIR_PERMISSIONS).Perm())
			if err != nil {
				ErrorAll("MAPPER: UNABLE TO CREATE MAP BACKUP DIRECTORY %s: %s", path, err.Error())
			}
		}

		obuff, err := os.ReadFile(path)
		if err != nil {
			ErrorAll("MAPPER: UNABLE TO OPEN MAP FILE %s: %s", path, err.Error())
		}

		nowTime := time.Now().Format("2006_01_02__15_04_05")
		backupFileName := fmt.Sprintf("%s/%s_%s_%s", backupDir, "backup", nowTime, filepath.Base(path))
		err = os.WriteFile(backupFileName, obuff, os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
		if err != nil {
			ErrorAll("MAPPER: CANNOT WRITE MAP FILE %s: %s", path, err)
		}

		Chown(backupDir, Config.WwwUser)
		Chown(backupFileName, Config.WwwUser)
	}

	var buff bytes.Buffer
	for _, e := range es {
		buff.WriteString(e.String())
	}

	err := os.WriteFile(path, buff.Bytes(), os.FileMode(OS_FILE_PERMISSIONS_R).Perm())
	if err != nil {
		ErrorAll("MAPPER: CANNOT WRITE MAP FILE %s: %s", path, err)
	}

	Chown(path, Config.WwwUser)
}

func Chown(path string, username string) {

	usr, err := user.Lookup(username)
	if err != nil {
		ErrorAll("MAPPER: CANNOT FIND USER %s TO GRANT WEB SERVER PERMISSIONS: %s", username, err)
	}

	uid, _ := strconv.Atoi(usr.Uid)
	gid, _ := strconv.Atoi(usr.Gid)

	err = os.Chown(path, uid, gid)
	if err != nil {
		ErrorAll("MAPPER: UNABLE TO GRANT WEB SERVER PERMISSIONS: %s", err)
	}
}
