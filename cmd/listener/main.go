package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"log"
	"wca-trapper/internal/config"
	"wca-trapper/internal/redis"
	"wca-trapper/internal/trap"
)

func main() {
	configPath := flag.String("config", "", "path to the configuration file")
	flag.Parse()

	conf := config.Configuration{}
	err := config.LoadConfig(*configPath, &conf)
	if err != nil {
		logrus.Fatal(err)
	}

	rds := redis.NewRedis(conf.Redis.Address, conf.Redis.Password, conf.Redis.Database, conf.Redis.Channel)
	var handler func(data interface{}) error
	if conf.Redis.Enabled {
		logrus.Info("sending to WCA enabled, try connect to redis")
		rds.TryConnect()
		handler = rds.Publish
	} else {
		handler = nil
	}
	logrus.Info("starting SNMP trap listener...")

	trapper := trap.NewTrapListener(conf.Listen.Address, conf.Listen.Community, false)
	trapper.SetPublisher(handler)
	err = trapper.Listen()
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}
