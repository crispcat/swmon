package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	TEMPLATE_NAGVIS_HEADER = ETC + "maps/nagvis-header.cfg"
	TEMPLATE_NAGVIS_HOST   = ETC + "maps/nagvis-host.cfg"
	TEMPLATE_NAGVIS_LINE   = ETC + "maps/nagvis-line.cfg"
)

//--------------------------------------------------------------------------------------------------------------------//

type NagvisMapEntityType byte

const (
	NagvisMapEntityHost NagvisMapEntityType = iota
	NagvisMapEntityLine
	NagvisMapEntityHeader
	NagvisMapEntityAsIs
)

const (
	NagvisMapEntityHostDefine = "host"
	NagvisMapEntityLineDefine = "servicegroup"
	NagvisMapHeaderDefine     = "global"
)

type NagvisMapEntity interface {
	Type() NagvisMapEntityType
	GetId() string
	String() string
	GetMapData() *NagvisMapEntityData
}

type NagvisMapEntityData struct {
	define       string
	etype        NagvisMapEntityType
	nagvisFields map[string]reflect.Value
	notParsed    string
}

func CreateEntityData(define string, etype NagvisMapEntityType) *NagvisMapEntityData {
	e := NagvisMapEntityData{
		define:       define,
		etype:        etype,
		nagvisFields: map[string]reflect.Value{},
	}
	return &e
}

func (d *NagvisMapEntityData) GetMapData() *NagvisMapEntityData {
	return d
}

func (d *NagvisMapEntityData) Type() NagvisMapEntityType {
	return d.etype
}

func (d *NagvisMapEntityData) String() string {

	var sb strings.Builder
	sb.WriteString("define")
	sb.WriteString(" ")
	sb.WriteString(d.define)
	sb.WriteString(" ")
	sb.WriteString("{")
	sb.WriteString("\n")
	switch d.etype {
	case NagvisMapEntityAsIs:
		sb.WriteString(d.notParsed)
	default:
		for n := range d.nagvisFields {
			sb.WriteString(n)
			sb.WriteString("=")
			val := d.Get(n)
			sb.WriteString(fmt.Sprintf("%v", val))
			sb.WriteString("\n")
		}
		sb.WriteString(d.notParsed)
	}
	sb.WriteString("}")
	sb.WriteString("\n")
	sb.WriteString("\n")
	return sb.String()
}

func (d *NagvisMapEntityData) GetId() string {

	switch d.etype {
	case NagvisMapEntityHost:
		idStr := d.Get("object_id")
		return idStr.(string)

	case NagvisMapEntityLine:
		return ""

	default:
		return ""
	}
}

func (d *NagvisMapEntityData) Set(name string, value string) bool {

	fields := d.nagvisFields
	p, ok := fields[name]
	if !ok {
		WriteAll("MAPPER: FIELD %s IS NOT DECLARED! FIELD WILL NOT BE CHENGED AND WILL PROCESSED AS IS", name)
		return false
	}

	Err := func(err error) {
		if err != nil {
			ErrorAll("MAPPER: CANNOT SET VALUE OF FIELD %s %s TO %s ", name, p.Type(), value)
		}
	}
	switch p.Kind() {
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		Err(err)
		p.SetInt(i)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		Err(err)
		p.SetBool(b)
	case reflect.String:
		p.SetString(value)
	default:
		ErrorAll("MAPPER: UNSUPPORTED TYPE %s PROVIDED FOR FIELD %s!", reflect.TypeOf(value), name)
	}

	return true
}

func (d *NagvisMapEntityData) Get(name string) any {

	fields := d.nagvisFields
	p, ok := fields[name]
	if !ok {
		WriteAll("MAPPER: FIELD %s IS NOT DECLARED!", name)
		return nil
	}

	switch p.Kind() {
	case reflect.Int64:
		return p.Int()
	case reflect.String:
		return p.String()
	case reflect.Bool:
		return p.Bool()
	default:
		ErrorAll("MAPPER: UNSUPPORTED TYPE %s PROVIDED FOR FIELD %s!", p.Kind(), name)
		return nil
	}
}

func MapFields(e any, fields map[string]reflect.Value) {

	r := reflect.Indirect(reflect.ValueOf(e))
	type_fields := reflect.VisibleFields(r.Type())
	for _, f := range type_fields {
		fname := f.Tag.Get("nagvis")
		if fname == "" {
			continue
		}
		fields[fname] = r.FieldByIndex(f.Index)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

type MapHost struct {
	Id          string `nagvis:"object_id"`
	X           int64  `nagvis:"x"`
	Y           int64  `nagvis:"y"`
	Z           int64  `nagvis:"z"`
	Name        string `nagvis:"host_name"`
	IconSet     string `nagvis:"iconset"`
	IconSize    int64  `nagvis:"icon_size"`
	ShowLabel   int64  `nagvis:"label_show"`
	LabelBg     string `nagvis:"label_background"`
	LabelBorder string `nagvis:"label_border"`
	LabelMaxlen int64  `nagvis:"label_maxlen"`
	*NagvisMapEntityData

	modelId uint32
}

type MapLine struct {
	Id               string `nagvis:"object_id"`
	X                string `nagvis:"x"`
	Y                string `nagvis:"y"`
	Z                int64  `nagvis:"z"`
	ViewType         string `nagvis:"view_type"`
	LineType         int64  `nagvis:"line_type"`
	ServiceGroupName string `nagvis:"servicegroup_name"`
	*NagvisMapEntityData
}

func CreateMapHost() *MapHost {
	h := &MapHost{}
	h.NagvisMapEntityData = CreateEntityData(NagvisMapEntityHostDefine, NagvisMapEntityHost)
	MapFields(h, h.nagvisFields)
	return h
}

func CopyMapHost(proto MapHost) *MapHost {
	h := &proto
	h.NagvisMapEntityData = CreateEntityData(NagvisMapEntityHostDefine, NagvisMapEntityHost)
	MapFields(h, h.nagvisFields)
	return h
}

func CreateMapLine() *MapLine {
	l := &MapLine{}
	l.NagvisMapEntityData = CreateEntityData(NagvisMapEntityLineDefine, NagvisMapEntityLine)
	MapFields(l, l.nagvisFields)
	return l
}

//--------------------------------------------------------------------------------------------------------------------//

func DrawMap(model *HostsModel, path string) {

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		CreateMap(model, path, false)
	} else {
		CreateMap(model, path, true)
	}
}

func CreateMap(model *HostsModel, path string, update bool) {

	WriteAll("MAPPER: CREATING NAGVIS MAP %s...", path)
	headerTemplateCfg, err := os.ReadFile(TEMPLATE_NAGVIS_HEADER)
	if err != nil {
		ErrorAll("MAPPER: UNABLE TO READ HEADER TEMPLATE FILE %s: %s", TEMPLATE_NAGVIS_HEADER, err)
	}
	chunks := strings.Fields(string(headerTemplateCfg))
	mapEntities := ParseMapEntities(chunks)

	hostTemplateCfg, err := os.ReadFile(TEMPLATE_NAGVIS_HOST)
	if err != nil {
		ErrorAll("MAPPER: UNABLE TO READ HOST TEMPLATE FILE %s: %s", TEMPLATE_NAGVIS_HOST, err)
	}

	pos := 1
	mapHostTemplate := CreateMapHost()
	chunks = strings.Fields(string(hostTemplateCfg))
	ParseMapEntity(chunks, &pos, mapHostTemplate)

	var readEntitiesById map[string]NagvisMapEntity
	if update {
		fileChunks := ReadMapFile(path)
		readEntities := ParseMapEntities(fileChunks)
		readEntitiesById = LookupMapEntities(readEntities)
	}

	mapHosts := map[uint32]*MapHost{}
	model.Map.Range(func(key, value any) bool {

		host := value.(*Host)
		if !host.WriteToConfig {
			return true
		}

		var mapHost *MapHost
		if readEntity, ok := readEntitiesById[host.MapId]; update && ok {
			mapHost = CopyMapHost(*readEntity.(*MapHost))
		} else {
			mapHost = CopyMapHost(*mapHostTemplate)
			mapHost.Id = RandomId()
		}

		host.MapId = mapHost.Id
		mapHost.Name = host.GetUniqueName()
		mapHost.modelId = host.Id
		mapHosts[host.Id] = mapHost

		MapFields(mapHost, mapHost.nagvisFields)
		mapEntities = append(mapEntities, mapHost)

		return true
	})

	lineTemplateCfg, err := os.ReadFile(TEMPLATE_NAGVIS_LINE)
	if err != nil {
		ErrorAll("MAPPER: UNABLE TO READ LINE TEMPLATE FILE %s: %s", TEMPLATE_NAGVIS_LINE, err)
	}

	pos = 1
	mapLineTemplate := CreateMapLine()
	chunks = strings.Fields(string(lineTemplateCfg))
	ParseMapEntity(chunks, &pos, mapLineTemplate)

	model.Conn.Range(func(key, value any) bool {

		conn := value.(Connection)

		frMapHost := mapHosts[conn.from]
		toMapHost := mapHosts[conn.to]

		mapLine := *mapLineTemplate
		mapLine.NagvisMapEntityData = CreateEntityData(NagvisMapEntityLineDefine, NagvisMapEntityLine)

		mapLine.Id = RandomId()
		mapLine.X = LinePosition(frMapHost.Id, toMapHost.Id)
		mapLine.Y = LinePosition(frMapHost.Id, toMapHost.Id)

		_frHost, _ := model.Map.Load(frMapHost.modelId)
		_toHost, _ := model.Map.Load(toMapHost.modelId)
		frHost := _frHost.(*Host)
		toHost := _toHost.(*Host)

		name1 := frHost.Ip.String()
		name2 := toHost.Ip.String()

		port1 := strconv.FormatUint(uint64(conn.frPort), 10)
		port2 := strconv.FormatUint(uint64(conn.toPort), 10)

		mapLine.ServiceGroupName = fmt.Sprintf(SERVICE_GROUP_FORMAT_STRING, name1, port1, name2, port2)

		MapFields(mapLine, mapLine.nagvisFields)
		mapEntities = append(mapEntities, mapLine)

		return true
	})

	WriteAll("MAPPER: CREATING NEW NAGVIS MAP ON PATH %s", path)
	WriteMapFile(path, mapEntities, update)
}

func ParseMapEntities(chunks []string) []NagvisMapEntity {

	var mapEntities []NagvisMapEntity
	var entity NagvisMapEntity
	for pos := 0; pos < len(chunks); pos++ {
		switch chunks[pos] {
		case "define":
			pos++
			switch chunks[pos] {
			case NagvisMapEntityHostDefine:
				entity = CreateMapHost()
			case NagvisMapEntityLineDefine:
				entity = CreateMapLine()
			default:
				entity = CreateEntityData(chunks[pos], NagvisMapEntityAsIs)
			}
		}
		ParseMapEntity(chunks, &pos, entity)
		mapEntities = append(mapEntities, entity)
	}
	return mapEntities
}

func ParseMapEntity(chunks []string, pos *int, e NagvisMapEntity) {

	*pos++
	if chunks[*pos] != "{" {
		SyntaxError("{", chunks, *pos)
	}

	entityData := e.GetMapData()

	*pos++
	bracketClosed := false
	for ; *pos < len(chunks); *pos++ {
		if chunks[*pos] == "}" {
			bracketClosed = true
			break
		}

		nameValue := strings.Split(chunks[*pos], "=")
		if len(nameValue) != 2 {
			SyntaxError("name=value", chunks, *pos)
		}
		name := strings.TrimFunc(nameValue[0], unicode.IsSpace)
		value := strings.TrimFunc(nameValue[1], unicode.IsSpace)

		if entityData.etype == NagvisMapEntityAsIs || !entityData.Set(name, value) {
			entityData.notParsed = entityData.notParsed + fmt.Sprintf("%s=%s\n", name, value)
		}
	}

	if !bracketClosed {
		SyntaxError("}", chunks, *pos)
	}
}

func SyntaxError(expected string, chunks []string, pos int) {
	ErrorAll("MAPPER: SYNTAX ERROR IN MAP CONFIG. EXPECTED %s AT %s", expected, PrintChunkArea(chunks, pos))
}

func PrintChunkArea(chunks []string, pos int) string {
	switch pos {
	case 0:
		return fmt.Sprintf("%s <<<< %s %s", chunks[pos], chunks[pos+1], chunks[pos+2])
	case len(chunks):
		return fmt.Sprintf("%s %s %s <<<<", chunks[pos-2], chunks[pos-1], chunks[pos])
	default:
		return fmt.Sprintf("%s %s <<<< %s ", chunks[pos-1], chunks[pos], chunks[pos+1])
	}
}

func LookupMapEntities(mapEntities []NagvisMapEntity) map[string]NagvisMapEntity {

	mapEntitesById := make(map[string]NagvisMapEntity, len(mapEntities))
	for _, entity := range mapEntities {
		entityId := entity.GetId()
		if entityId != "" {
			mapEntitesById[entityId] = entity
		}
	}

	return mapEntitesById
}

func RandomId() string {
	rnd := rand.Uint32()
	return hex.EncodeToString([]byte{byte(rnd), byte(rnd >> 8), byte(rnd >> 16)})
}

func LinePosition(fromId string, toId string) string {
	return fmt.Sprintf("%s%%+11,%s%%+11", fromId, toId)
}
