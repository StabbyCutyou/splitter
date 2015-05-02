package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/StabbyCutyou/splitter/config"
	"github.com/StabbyCutyou/splitter/server"
)

const VERSION = "0.0.1"

func main() {
	logrus.Infof("Splitter version %s booting...", VERSION)
	// pre-config
	config_file := flag.String("c", "./config/splitter.cfg", "location of config file")
	flag.Parse()

	cfg, err := config.GetConfig(config_file)
	if err != nil {
		logrus.Error(err)
	}
	// Begin listening

	// main thread now blocked
	server.StartReadListening(cfg.Network.ListenerPort, cfg.Network.WriterPort, BumpBytes)
}

func BumpBytes(bytes []byte) []byte {
	for i, b := range bytes {
		bytes[i] = b + byte(1)
	}
	return bytes
}
