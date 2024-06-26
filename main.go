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

func BuildBackupProvider(conf *BackupProviderConfig) (BackupProvider, error) {
	switch conf.Type {
	case BackupProviderConfigTypeGitea:
		c, err := ConvertToBackupProviderConfig[GiteaConf](conf.Config)
		if err != nil {
			return nil, err
		}
		return NewGitea(c), nil
	case BackupProviderConfigTypeFile:
		c, err := ConvertToBackupProviderConfig[FileBackupConfig](conf.Config)
		if err != nil {
			log.Fatalf("convert file config error: %s", err.Error())
		}
		return NewFileBackup(c)
	}
	return nil, fmt.Errorf("unknown backup provider type: %s", conf.Type)
}

func runBackupTask(conf *SyncConfig) {
	for _, target := range conf.Targets {

		// merge default config
		target.MergeDefault(conf.DefaultConf)

		// load all github repos
		github := NewGithub(target.Token)
		repos, err := github.LoadAllRepos(target.Owner, target.IsOwnerOrg)
		if err != nil {
			log.Panicf("load %s repos error: %s", target.RepoOwner, err.Error())
		}

		// build backup provider
		provider, err := BuildBackupProvider(target.Backup)
		if err != nil {
			log.Panicf("build backup provider error: %s", err.Error())
		}

		// handle repos set
		handledRepos := make(map[string]struct{})

		for _, repo := range repos {
			// render repo identity
			identity := RepoIdentity(target.Owner, repo.Name, repo.Private, repo.Fork, repo.Archived)

			// check allow/deny rule
			if target.Filter != nil {
				if !IsMatchRepoIdentity(identity, target.Filter.AllowRule...) {
					if IsMatchRepoIdentity(identity, target.Filter.DenyRule...) {
						continue
					}
				}
			}

			githubToken := target.Token
			// check specific GitHub token for this repo by regex
			for k, v := range target.SpecificGithubToken {
				if IsMatchRepoIdentity(identity, k) {
					githubToken = v
					break
				}
			}

			// migrate repo
			s, e := provider.MigrateRepo(
				target.Owner, target.RepoOwner,
				target.IsOwnerOrg, target.IsRepoOwnerOrg,
				repo.Name, repo.Description,
				githubToken,
			)

			if e != nil {
				log.Printf("migrate %s error: %s", repo.Name, e.Error())
			} else {
				log.Printf("migrate %s %s", repo.Name, s)
			}
			handledRepos[repo.Name] = struct{}{}
		}

		// delete unmatched repos if needed
		if target.Filter.UnmatchedRepoAction == UnmatchedRepoActionDelete {

			// load local repos
			localRepos, lErr := provider.LoadRepos(target.RepoOwner, target.IsRepoOwnerOrg)
			if lErr != nil {
				log.Panicf("load %s repos error: %s", target.RepoOwner, lErr.Error())
			}

			// delete unmatched repos
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
