package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"isPrivate"`
	Fork        bool   `json:"isFork"`
	Archived    bool   `json:"isArchived"`
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

func (g *Github) LoadAllRepos(owner string, isOrg bool) ([]Repo, error) {
	tmpl := `
query {
  %s {
    repositories(
      first: 100,
      after: %s
    ) {
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        name
        description
        isPrivate
        isFork
	    isArchived
        owner {
          login
        }
      }
    }
  }
}
`
	next := "null"
	queryType := ""
	dataKey := ""
	var repos []Repo
	if isOrg {
		dataKey = "organization"
		queryType = fmt.Sprintf("organization(login: \"%s\")", owner)
	} else {
		dataKey = "repositoryOwner"
		queryType = fmt.Sprintf("repositoryOwner(login: \"%s\")", owner)
	}
	for {
		var data struct {
			Data map[string]struct {
				Repositories struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []Repo `json:"nodes"`
				} `json:"repositories"`
			} `json:"data"`
		}
		query := fmt.Sprintf(tmpl, queryType, next)
		rawBody, err := json.Marshal(map[string]any{"query": query})
		if err != nil {
			return nil, err
		}
		err = GithubRequestJson("POST", "graphql", g.Token, bytes.NewReader(rawBody), &data)
		if err != nil {
			return nil, err
		}
		info, ok := data.Data[dataKey]
		if !ok {
			return nil, fmt.Errorf("no data key %s", dataKey)
		}
		for _, repo := range info.Repositories.Nodes {
			if repo.Owner.Login == owner {
				repos = append(repos, repo)
			}
		}
		if !info.Repositories.PageInfo.HasNextPage {
			break
		}
		next = fmt.Sprintf(`"%s"`, info.Repositories.PageInfo.EndCursor)
	}
	return repos, nil
}

func GithubRequest(method, path, token string, body io.Reader, handler func(*http.Response) error) error {
	url := fmt.Sprintf("https://api.github.com/%s", path)
	req, err := http.NewRequest(method, url, body)
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s error: %s", path, resp.Status)
	}
	return handler(resp)
}

func GithubRequestJson[T any](method, path, token string, body io.Reader, res *T) error {
	return GithubRequest(method, path, token, body, func(resp *http.Response) error {
		return json.NewDecoder(resp.Body).Decode(res)
	})
}
