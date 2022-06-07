package main

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var Cfg struct {
	Debug   bool
	Account []struct {
		Name     string
		Password string
	}
	App struct {
		Https     bool
		Port      int
		Log       string
		Resource  string
		Signature struct {
			Crt string
			Key string
		}
	}
	// Codecs struct{

	// }
}

func init() {
	var config_path string
	flag.StringVar(&config_path, "c", "conf.yml", "set path to configuration file")
	flag.Parse()

	if conf_file, err := ioutil.ReadFile(config_path); err != nil {
		panic("Cannot open file: " + config_path + ".\n" + err.Error())
	} else if err := yaml.Unmarshal(conf_file, &Cfg); err != nil {
		panic("File: " + config_path + " error.\n" + err.Error())
	}
}
