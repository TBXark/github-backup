package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	BuildVersion = "dev"
)

type BackupProvider interface {
	LoadRepos(owner string, isOrg bool) ([]string, error)
	MigrateRepo(
		owner, repoOwner string,
		isOwnerOrg, isRepoOwnerOrg bool,
		repoName, repoDesc string,
		githubToken string,
	) (string, error)
	DeleteRepo(owner, repo string) (string, error)
}

func BackupProviderBuilder(conf *BackupProviderConfig) BackupProvider {
	switch conf.Type {
	case BackupProviderConfigTypeGitea:
		c, err := ConvertToBackupProviderConfig[GiteaConf](conf.Config)
		if err != nil {
			log.Fatalf("convert gitea config error: %s", err.Error())
		}
		return NewGitea(c)
	case BackupProviderConfigTypeFile:
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
		repos, tErr := github.LoadAllRepos(target.Owner, target.IsOwnerOrg)
		if tErr != nil {
			log.Panicf("load %s repos error: %s", target.RepoOwner, tErr.Error())
		}

		provider := BackupProviderBuilder(target.Backup)
		handledRepos := make(map[string]struct{})

		for _, repo := range repos {
			desc := RepoDescription(target.Owner, repo.Name, repo.Private, repo.Fork, repo.Archived)
			if !IsMatchRepoDescription(desc, target.Filter.AllowRule...) {
				if IsMatchRepoDescription(desc, target.Filter.DenyRule...) {
					continue
				}
			}
			s, e := provider.MigrateRepo(
				target.Owner, target.RepoOwner,
				target.IsOwnerOrg, target.IsRepoOwnerOrg,
				repo.Name, repo.Description,
				target.Token,
			)
			if e != nil {
				log.Printf("migrate %s error: %s", repo.Name, e.Error())
			} else {
				log.Printf("migrate %s %s", repo.Name, s)
			}
			handledRepos[repo.Name] = struct{}{}
		}

		if target.Filter.UnmatchedRepoAction == UnmatchedRepoActionDelete {
			localRepos, lErr := provider.LoadRepos(target.RepoOwner, target.IsRepoOwnerOrg)
			if lErr != nil {
				log.Panicf("load %s repos error: %s", target.RepoOwner, lErr.Error())
			}
			for _, repo := range localRepos {
				if _, ok := handledRepos[repo]; !ok {
					s, e := provider.DeleteRepo(target.RepoOwner, repo)
					if e != nil {
						log.Printf("delete %s error: %s", repo, e.Error())
					} else {
						log.Printf("delete %s %s", repo, s)
					}
				}
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
