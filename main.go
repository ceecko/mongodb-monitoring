package main

import (
	"context"
	"fmt"

	"github.com/ceecko/mongodb-monitoring/config"
	"github.com/ceecko/mongodb-monitoring/monitoring"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	configPath = pflag.String("config", "./config.yaml", "Path to a config file")
	debug      = pflag.Bool("debug", false, "Enable debug logging")
)

func main() {
	fmt.Printf("MongoDB Monitoring version %s\n", Version)

	pflag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		logrus.Panic(err)
	}

	m := monitoring.NewMonitoring(cfg)
	if err = m.Start(context.Background()); err != nil {
		logrus.Panic(err)
	}
}
