package main

import (
	"encoding/json"
	"os"
)

type BackupProviderConfigType string

const (
	BackupProviderConfigTypeGitea BackupProviderConfigType = "gitea"
	BackupProviderConfigTypeFile  BackupProviderConfigType = "file"
)

type UnmatchedRepoAction string

const (
	UnmatchedRepoActionDelete UnmatchedRepoAction = "delete"
	UnmatchedRepoActionIgnore UnmatchedRepoAction = "ignore"
)

type BackupProviderConfig struct {
	Type   BackupProviderConfigType `json:"type"`
	Config any                      `json:"config"`
}

type DefaultConfig struct {
	GithubToken string                `json:"github_token"`
	RepoOwner   string                `json:"repo_owner"`
	Backup      *BackupProviderConfig `json:"backup"`
	Filter      *FilterConfig         `json:"filter"`
}

type GithubConfig struct {
	Owner          string                `json:"owner"`
	Token          string                `json:"token"`
	IsOwnerOrg     bool                  `json:"is_owner_org"`
	RepoOwner      string                `json:"repo_owner"`
	IsRepoOwnerOrg bool                  `json:"is_repo_owner_org"`
	Backup         *BackupProviderConfig `json:"backup"`
	Filter         *FilterConfig         `json:"filter"`
}

type FilterConfig struct {
	UnmatchedRepoAction UnmatchedRepoAction `json:"unmatched_repo_action"`
	AllowRule           []string            `json:"allow_rule"`
	DenyRule            []string            `json:"deny_rule"`
}

func (c *GithubConfig) MergeDefault(defaultConf *DefaultConfig) {
	if defaultConf == nil {
		return
	}
	if c.Token == "" {
		c.Token = defaultConf.GithubToken
	}
	if c.RepoOwner == "" {
		c.RepoOwner = defaultConf.RepoOwner
	}
	if c.Backup == nil {
		c.Backup = defaultConf.Backup
	}
	if c.Filter == nil {
		c.Filter = defaultConf.Filter
		if c.Filter == nil {
			c.Filter = &FilterConfig{}
		}
	}
	if c.Filter.UnmatchedRepoAction == "" {
		c.Filter.UnmatchedRepoAction = defaultConf.Filter.UnmatchedRepoAction
		if c.Filter.UnmatchedRepoAction == "" {
			c.Filter.UnmatchedRepoAction = UnmatchedRepoActionIgnore
		}
	}
}

type SyncConfig struct {
	DefaultConf *DefaultConfig `json:"default_conf"`
	Targets     []GithubConfig `json:"targets"`
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
