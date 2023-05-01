package main

import (
	"fmt"
	"log"
	"os"
)

func WriteAll(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
	log.Printf(format+"\n", v...)
}

func ErrorAll(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
	log.Fatalf(format+"\n", v...)
}

func Log(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

func Stdout(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}

func Verbose(format string, v ...any) {
	if IsVerbose {
		WriteAll(format+"\n", v...)
	}
}

var LogsFile *os.File

func DoLogs(path string) error {

	var err error
	LogsFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(LogsFile)
	return nil
}

func DoneLogs() {

	_ = LogsFile.Close()
}
