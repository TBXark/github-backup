package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Github struct {
	Token string
}

func NewGithub(token string) *Github {
	return &Github{Token: token}
}

func (g *Github) loadRepos(owner string, perPage, page int) ([]Repo, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=%d&page=%d", owner, perPage, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.Token))
	}
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

func (g *Github) LoadRepos(owner string) ([]Repo, error) {
	perPage := 100
	page := 1
	res := make([]Repo, 0)
	for {
		repos, err := g.loadRepos(owner, perPage, page)
		if err != nil {
			break
		}
		if len(repos) == 0 {
			break
		}
		res = append(res, repos...)
		page++
	}
	return res, nil
}

func GithubRequest(method, path, token string, handler func(*http.Response) error) error {
	url := fmt.Sprintf("https://api.github.com/%s", path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s error: %s", path, resp.Status)
	}
	defer resp.Body.Close()
	return handler(resp)
}

func GithubRequestJson[T any](method, path, token string, res *T) error {
	return GithubRequest(method, path, token, func(resp *http.Response) error {
		return json.NewDecoder(resp.Body).Decode(res)
	})
}
