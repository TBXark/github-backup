package main

import (
	"encoding/json"
	"os"
)

type BackupProviderConfig struct {
	Type   string `json:"type"`
	Config any    `json:"config"`
}

type DefaultConfig struct {
	GithubToken string                `json:"github_token"`
	RepoOwner   string                `json:"repo_owner"`
	Backup      *BackupProviderConfig `json:"backup"`
}

type GithubConfig struct {
	Owner     string                `json:"owner"`
	Token     string                `json:"token"`
	RepoOwner string                `json:"repo_owner"`
	Backup    *BackupProviderConfig `json:"backup"`
}

func (c *GithubConfig) MergeDefault(defaultConf *DefaultConfig) {
	if c.Token == "" {
		c.Token = defaultConf.GithubToken
	}
	if c.RepoOwner == "" {
		c.RepoOwner = defaultConf.RepoOwner
	}
	if c.Backup == nil {
		c.Backup = defaultConf.Backup
	}
}

type SyncConfig struct {
	Host         string         `json:"host"`
	GiteaToken   string         `json:"gitea_token"`
	AuthUsername string         `json:"auth_username"`
	DefaultConf  *DefaultConfig `json:"default_conf"`
	Targets      []GithubConfig `json:"targets"`
}

func ConvertToBackupProviderConfig[T any](raw any) (*T, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var conf T
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func LoadConfig(path string) (*SyncConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf SyncConfig
	err = json.Unmarshal(file, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
