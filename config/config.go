package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	GithubToken         string                `json:"github_token"`
	RepoOwner           string                `json:"repo_owner"`
	Backup              *BackupProviderConfig `json:"backup"`
	Filter              *FilterConfig         `json:"filter"`
	SpecificGithubToken map[string]string     `json:"specific_github_token"`
}

type GithubConfig struct {
	Owner               string                `json:"owner"`
	Token               string                `json:"token"`
	IsOwnerOrg          bool                  `json:"is_owner_org"`
	RepoOwner           string                `json:"repo_owner"`
	IsRepoOwnerOrg      bool                  `json:"is_repo_owner_org"`
	Backup              *BackupProviderConfig `json:"backup"`
	Filter              *FilterConfig         `json:"filter"`
	SpecificGithubToken map[string]string     `json:"specific_github_token"`
}

type FilterConfig struct {
	UnmatchedRepoAction UnmatchedRepoAction `json:"unmatched_repo_action"`
	PreDeleteCheckCount int                 `json:"pre_delete_check_count"`
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
	if c.RepoOwner == "" {
		c.RepoOwner = c.Owner
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
		c.Filter.PreDeleteCheckCount = defaultConf.Filter.PreDeleteCheckCount
		if c.Filter.UnmatchedRepoAction == "" {
			c.Filter.UnmatchedRepoAction = UnmatchedRepoActionIgnore
		}
	}
	if len(c.Filter.AllowRule) == 0 {
		c.Filter.AllowRule = defaultConf.Filter.AllowRule
	}
	if len(c.Filter.DenyRule) == 0 {
		c.Filter.DenyRule = defaultConf.Filter.DenyRule
	}
	if len(c.SpecificGithubToken) == 0 {
		c.SpecificGithubToken = defaultConf.SpecificGithubToken
	}
}

type SyncConfig struct {
	DefaultConf *DefaultConfig  `json:"default_conf"`
	Targets     []*GithubConfig `json:"targets"`
	Cron        string          `json:"cron"`
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
	var body []byte
	var err error

	if strings.HasPrefix(path, "http") {
		resp, httpErr := http.Get(path)
		if httpErr != nil {
			return nil, fmt.Errorf("failed to fetch config: %w", httpErr)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
	} else {
		body, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var conf SyncConfig
	err = json.Unmarshal(body, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
