package github

import (
	"fmt"
	"github.com/TBXark/github-backup/utils/request"
	"strings"
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
	token := request.WithAuthorization(g.Token, "bearer")
	ownerLower := strings.ToLower(owner)
	for {
		query := map[string]string{"query": fmt.Sprintf(tmpl, queryType, next)}
		data, err := request.POST[reposQuery]("https://api.github.com/graphql", query, token)
		if err != nil {
			return nil, err
		}
		info, ok := data.Data[dataKey]
		if !ok {
			return nil, fmt.Errorf("no data key %s", dataKey)
		}
		for _, repo := range info.Repositories.Nodes {
			if strings.ToLower(repo.Owner.Login) == ownerLower {
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

type reposQuery struct {
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
