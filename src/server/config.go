package main

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var Cfg struct {
	Debug bool
	App   struct {
		Resource string
		Port     int
	}
}

func init() {
	var config_path string
	flag.StringVar(&config_path, "c", "conf.yml", "set path to configuration file")
	flag.Parse()

	conf_file, err := ioutil.ReadFile(config_path)
	if err != nil {
		panic(err.Error())
	}

	err = yaml.Unmarshal(conf_file, &Cfg)
	if err != nil {
		panic(err.Error())
	}
}
