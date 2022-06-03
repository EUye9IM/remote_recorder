package main

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var Cfg struct {
	Debug bool
	App   struct {
		Https     bool
		Port      int
		Resource  string
		Sign_path struct {
			Crt string
			Key string
		}
	}
	Webrtc struct {
		Ice_list []string
	}
}

func init() {
	var config_path string
	flag.StringVar(&config_path, "c", "conf.yml", "set path to configuration file")
	flag.Parse()

	if conf_file, err := ioutil.ReadFile(config_path); err != nil {
		panic(err.Error())
	} else if err := yaml.Unmarshal(conf_file, &Cfg); err != nil {
		panic(err.Error())
	}
}
