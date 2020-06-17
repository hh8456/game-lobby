package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type SConfig struct {
	RdsAddr    string `yaml:"redis_addr"`
	MysqlAddr  string `yaml:"mysql_addr"`
	LobbyAddr  string `yaml:"lobby_addr"`
	NiuNIuAddr string `yaml:"niuniu_addr"`
}

var Cfg *SConfig = nil

func InitSConfig(strConfig string) error {
	Cfg = &SConfig{}

	file, err := os.Open(strConfig)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, Cfg); err != nil {
		return err
	}

	return nil
}
