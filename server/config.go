package main

import (
	"log"

	"gopkg.in/ini.v1"
)

//////////////////////   CONFIG
type TConfig struct {
	Server TMySQLSettings
}

type TMySQLSettings struct {
	SrcUserName string
	SrcPassword string
	SrcNetwork  string
	SrcAddress  string
	DBName      string
	Charset     string
}

func LoadConfig(iniFilepath string) (TConfig, error) {
	var c TConfig
	cfg, err := ini.Load(iniFilepath)

	if err != nil {
		log.Println(iniFilepath, err)
		return c, err
	}

	c.Server.SrcUserName = cfg.Section("DBServer").Key("SrcUserName").String()
	c.Server.SrcPassword = cfg.Section("DBServer").Key("SrcPassword").String()
	c.Server.SrcNetwork = cfg.Section("DBServer").Key("SrcNetwork").String()
	c.Server.SrcAddress = cfg.Section("DBServer").Key("SrcAddress").String()
	c.Server.DBName = cfg.Section("DBServer").Key("db_name").String()
	c.Server.Charset = cfg.Section("DBServer").Key("charset").String()

	return c, err
}
