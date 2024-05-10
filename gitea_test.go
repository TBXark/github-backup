package main

import "testing"

func TestGitea_DeleteAllRepos(t *testing.T) {
	conf, err := LoadConfig("./config.json")
	if err != nil {
		t.Error(err)
		return
	}
	provider, err := BuildBackupProvider(conf.DefaultConf.Backup)
	if err != nil {
		t.Error(err)
		return
	}
	gitea := provider.(*Gitea)
	for _, target := range conf.Targets {
		gitea.DeleteAllRepos(target.Owner, target.IsOwnerOrg)
	}

}
