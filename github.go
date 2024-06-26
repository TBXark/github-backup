package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
	Owner       struct {
		Login string `json:"login"`
	}
}

type Github struct {
	Token string
}

func NewGithub(token string) *Github {
	return &Github{Token: token}
}

func filterRepos(owner string, res []Repo) map[string]Repo {
	matchedRepos := make(map[string]Repo)
	ownerLower := strings.ToLower(owner)
	for _, repo := range res {
		if strings.ToLower(repo.Owner.Login) != ownerLower {
			continue
		}
		matchedRepos[repo.Name] = repo
	}
	return matchedRepos
}

func (g *Github) loadReposPageBySearch(owner string, perPage, page int, isOrg bool) (map[string]Repo, error) {
	url := "search/repositories"
	if isOrg {
		url = fmt.Sprintf("%s?q=org%s&page=%d&per_page=%d", url, "%3A"+owner, page, perPage)
	} else {
		url = fmt.Sprintf("%s?q=user%s&page=%d&per_page=%d", url, "%3A"+owner, page, perPage)
	}
	var res struct {
		Items []Repo `json:"items"`
	}
	err := GithubRequestJson("GET", url, g.Token, &res)
	if err != nil {
		return nil, err
	}
	return filterRepos(owner, res.Items), nil

}

func (g *Github) loadReposPage(owner string, perPage, page int, isOrg bool) (map[string]Repo, error) {
	url := fmt.Sprintf("users/%s/repos", owner)
	if isOrg {
		url = fmt.Sprintf("orgs/%s/repos", owner)
	}
	url = fmt.Sprintf("%s?per_page=%d&page=%d&type=all", url, perPage, page)
	var res []Repo
	err := GithubRequestJson("GET", url, g.Token, &res)
	if err != nil {
		return nil, err
	}
	return filterRepos(owner, res), nil
}

func (g *Github) LoadAllRepos(owner string, isOrg bool) ([]Repo, error) {
	page := 1
	perPage := 100
	res := make(map[string]Repo)
	for {
		repos, err := g.loadReposPageBySearch(owner, perPage, page, isOrg)
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		for k, v := range repos {
			res[k] = v
		}
		page++
	}
	page = 1
	for {
		repos, err := g.loadReposPage(owner, perPage, page, isOrg)
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		for k, v := range repos {
			res[k] = v
		}
		page++
	}
	var resSlice []Repo
	for _, v := range res {
		resSlice = append(resSlice, v)
	}
	return resSlice, nil
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
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s error: %s", path, resp.Status)
	}
	return handler(resp)
}

func GithubRequestJson[T any](method, path, token string, res *T) error {
	return GithubRequest(method, path, token, func(resp *http.Response) error {
		return json.NewDecoder(resp.Body).Decode(res)
	})
}
