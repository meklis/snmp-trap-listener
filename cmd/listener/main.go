package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"log"
	"wca-trapper/internal/config"
	"wca-trapper/internal/pubs/redis"
	"wca-trapper/internal/pubs/script_handler"
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

	trapper := trap.NewTrapListener(conf.Listen.Address, conf.Listen.Community, false)

	if conf.Redis.Enabled {
		logrus.Info("sending to redis enabled, try connect to redis")
		rds := redis.NewRedis(conf.Redis.Address, conf.Redis.Password, conf.Redis.Database, conf.Redis.Channel)
		rds.TryConnect()
		trapper.AddPublisher(rds.Publish)
	}
	if conf.ScriptHandler.Enabled {
		logrus.Info("sending to handler enabled, try connect to redis")
		handler := script_handler.NewScriptHandler(conf.ScriptHandler.Command, conf.ScriptHandler.CountHandlers, conf.ScriptHandler.QueueSize)
		trapper.AddPublisher(handler.Publish)
	}

	logrus.Info("starting SNMP trap listener...")
	err = trapper.Listen()
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}
