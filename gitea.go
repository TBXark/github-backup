package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type giteaMigrateRequest struct {
	Description    string `json:"description"`
	Private        bool   `json:"private"`
	PullRequests   bool   `json:"pull_requests"`
	Uid            int    `json:"uid"`
	AuthUsername   string `json:"auth_username"`
	AuthToken      string `json:"auth_token"`
	Issues         bool   `json:"issues"`
	Labels         bool   `json:"labels"`
	Milestones     bool   `json:"milestones"`
	Wiki           bool   `json:"wiki"`
	Releases       bool   `json:"releases"`
	MirrorInterval string `json:"mirror_interval"`
	RepoOwner      string `json:"repo_owner"`
	Service        string `json:"service"`
	RepoName       string `json:"repo_name"`
	CloneAddr      string `json:"clone_addr"`
	Mirror         bool   `json:"mirror"`
	Lfs            bool   `json:"lfs"`
}

type GiteaConf struct {
	Host         string `json:"host"`
	Token        string `json:"token"`
	AuthUsername string `json:"auth_username"`
}

func NewGiteaConf(raw any) *GiteaConf {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var conf GiteaConf
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return nil
	}
	return &conf
}

type Gitea struct {
	conf *GiteaConf
}

func NewGitea(conf *GiteaConf) *Gitea {
	return &Gitea{conf: conf}
}

func (g *Gitea) LoadRepos(owner string) ([]Repo, error) {
	url := fmt.Sprintf("%s/api/v1/user/repos", g.conf.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", g.conf.Token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var repos []Repo
	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (g *Gitea) MigrateRepo(owner, repoOwner, githubToken string, repo Repo) (string, error) {
	r := giteaMigrateRequest{
		Description:    repo.Description,
		Private:        true,
		PullRequests:   false,
		Uid:            0,
		AuthUsername:   g.conf.AuthUsername,
		AuthToken:      githubToken,
		Issues:         false,
		Labels:         false,
		Milestones:     false,
		Wiki:           false,
		Releases:       false,
		MirrorInterval: "10m0s",
		RepoOwner:      repoOwner,
		Service:        "github",
		RepoName:       repo.Name,
		CloneAddr:      fmt.Sprintf("https://github.com/%s/%s.git", owner, repo.Name),
		Mirror:         true,
		Lfs:            false,
	}
	rb, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	body := bytes.NewBuffer(rb)

	url := fmt.Sprintf("%s/api/v1/repos/migrate", g.conf.Host)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", g.conf.Host))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

func (g *Gitea) DeleteRepo(owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s", g.conf.Host, owner, repo)
	fmt.Println(url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", g.conf.Token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

func (g *Gitea) DeleteAllRepos(owner string) {
	for {
		repos, err := g.LoadRepos(owner)
		if err != nil {
			log.Panicf("get all repos error: %e", err)
		}
		if len(repos) == 0 {
			break
		}
		for _, repo := range repos {
			resp, e := g.DeleteRepo(owner, repo.Name)
			if e != nil {
				log.Printf("delete %s error: %e", repo.Name, e)
			} else {
				log.Printf("delete %s success: %s", repo.Name, resp)
			}
		}
	}
}
