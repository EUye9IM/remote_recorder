package main

import (
	"flag"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

var config struct {
	Debug     bool
	Save_path string
	Database  struct {
		Path    string
		Initial bool
		Account []struct {
			No       string
			Name     string
			Password string
			Level    string
			Enable   string
		}
	}
	Password struct {
		Strict_mode bool
		Length      int
		Upper       int
		Lower       int
		Digital     int
		Other       int
	}
	App struct {
		Port      int
		Https     bool
		Log       string
		Resource  string
		Signature struct {
			Crt string
			Key string
		}
	}
}

func init() {
	var config_path string
	flag.StringVar(&config_path, "c", "webrtc-conf.toml", "set path to configuration file")
	flag.Parse()

	if _, err := toml.DecodeFile(config_path, &config); err != nil {
		panic("File: " + config_path + " error.\n" + err.Error())
	}
	if !strings.HasSuffix(config.Save_path, "/") {
		config.Save_path += "/"
	}
	os.MkdirAll(config.Save_path, os.ModePerm)
	dbinit()
}
