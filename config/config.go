package config

import (
	"code.google.com/p/gcfg"
	"github.com/Sirupsen/logrus"
	"strings"
)

type Config struct {
	Main    Main
	Network Network
	Writing Writing
}

type Main struct {
	Name string
}

type Network struct {
	ListenerPort        int
	WriterPort          int
	InitialListenerPool int
}

type Writing struct {
	WriteTo     string
	WriteToList []string
}

func GetConfig(file_name *string) (*Config, error) {
	if *file_name == "" {
		logrus.Warn("Empty value provided for config file location from flag -c : Falling back to default location './lib/config.gcfg'")
		*file_name = "./config/splitter.cfg"
	}
	logrus.Info("Using config file located at ", *file_name)
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, *file_name)
	cfg.Writing.WriteToList = make([]string, 0)
	if cfg.Writing.WriteTo != "" {
		cfg.Writing.WriteToList = strings.Split(cfg.Writing.WriteTo, ",")
	}
	if err != nil {
		logrus.Fatal(err)
	}
	return &cfg, err
}
