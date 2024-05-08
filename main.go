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
	LoadRepos(owner string) ([]string, error)
	MigrateRepo(owner, repoOwner string, repoName, repoDesc string, githubToken string) (string, error)
	DeleteRepo(owner, repo string) (string, error)
}

func BackupProviderBuilder(conf *BackupProviderConfig) BackupProvider {
	switch conf.Type {
	case "gitea":
		c, err := ConvertToBackupProviderConfig[GiteaConf](conf.Config)
		if err != nil {
			log.Fatalf("convert gitea config error: %s", err.Error())
		}
		return NewGitea(c)
	case "file":
		c, err := ConvertToBackupProviderConfig[FileBackupConfig](conf.Config)
		if err != nil {
			log.Fatalf("convert file config error: %s", err.Error())
		}
		return NewFileBackup(c)
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
			s, e := provider.MigrateRepo(
				target.Owner, target.RepoOwner,
				repo.Name, repo.Description,
				target.Token,
			)
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
