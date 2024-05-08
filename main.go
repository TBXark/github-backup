package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	BuildVersion = "dev"
)

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BackupProvider interface {
	LoadRepos(owner string) ([]Repo, error)
	MigrateRepo(owner, repoOwner, githubToken string, repo Repo) (string, error)
	DeleteRepo(owner, repo string) (string, error)
}

func BackupProviderBuilder(conf *BackupProviderConfig) BackupProvider {
	switch conf.Type {
	case "gitea":
		c := NewGiteaConf(conf.Config)
		if c == nil {
			return nil
		}
		return NewGitea(c)
	}
	return nil

}

func runBackupTask(conf *SyncConfig) {
	for _, target := range conf.Targets {
		target.MergeDefault(conf.DefaultConf)
		github := NewGithub(target.Token)
		repos, tErr := github.LoadRepos(target.Owner)
		if tErr != nil {
			log.Panicf("load %s repos error: %s", target.RepoOwner, tErr.Error())
		}
		provider := BackupProviderBuilder(target.Backup)
		for _, repo := range repos {
			s, e := provider.MigrateRepo(target.Owner, target.RepoOwner, target.Token, repo)
			if e != nil {
				log.Printf("migrate %s error: %s", repo.Name, e.Error())
			} else {
				log.Printf("migrate %s %s", repo.Name, s)
			}
		}
	}
}

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
	conf, err := LoadConfig(*c)
	if err != nil {
		log.Fatalf("load config error: %s", err.Error())
	}
	runBackupTask(conf)
}
