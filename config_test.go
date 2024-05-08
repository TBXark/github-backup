package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSyncConfig(t *testing.T) {
	c := SyncConfig{
		Host:         "",
		GiteaToken:   "",
		AuthUsername: "",
		DefaultConf: &DefaultConfig{
			GithubToken: "",
			RepoOwner:   "",
			Backup: &BackupProviderConfig{
				Type: "",
				Config: &GiteaConf{
					Host:         "a",
					Token:        "b",
					AuthUsername: "c",
				},
			},
		},
		Targets: []GithubConfig{
			{
				Owner:     "",
				Token:     "",
				RepoOwner: "",
				Backup:    nil,
			},
		},
	}
	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", c)
	giteaConf := NewGiteaConf(c.DefaultConf.Backup.Config)
	if giteaConf == nil {
		t.Fatal("NewGiteaConf error")
	}
	if giteaConf.Host != "a" {
		t.Fatal("NewGiteaConf error")
	}
	if giteaConf.Token != "b" {
		t.Fatal("NewGiteaConf error")
	}
	if giteaConf.AuthUsername != "c" {
		t.Fatal("NewGiteaConf error")
	}
}
