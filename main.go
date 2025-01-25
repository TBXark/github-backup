package main

import (
	"flag"
	"fmt"
	"github.com/TBXark/confstore"
	"github.com/TBXark/github-backup/config"
	"github.com/robfig/cron/v3"
	"log"
)

var (
	BuildVersion = "dev"
)

func main() {
	c := flag.String("config", "config.json", "config file")
	v := flag.Bool("version", false, "show version")
	h := flag.Bool("help", false, "show help")
	flag.Parse()
	if *v {
		fmt.Println(BuildVersion)
		return
	}
	if *h {
		flag.Usage()
		return
	}
	conf, err := confstore.Load[config.SyncConfig](*c)
	if err != nil {
		log.Fatalf("load config error: %s", err.Error())
	}

	syncTask := NewTask(conf)
	if conf.Cron != "" {
		task := cron.New()
		_, e := task.AddJob(conf.Cron, syncTask)
		if e != nil {
			log.Fatalf("add cron task error: %s", e.Error())
		}
		task.Run()
	} else {
		syncTask.Run()
	}
}
