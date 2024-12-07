package main

import (
	"os"
	"testing"
)

func TestGithub_LoadAllRepos(t *testing.T) {
	github := NewGithub(os.Getenv("GITHUB_TOKEN"))
	repos, err := github.LoadAllRepos("TBXark", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("repos count: %d", len(repos))
	for _, repo := range repos {
		t.Log(repo.Name)
	}

	repos, err = github.LoadAllRepos("tbxark-arc", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("repos count: %d", len(repos))
	for _, repo := range repos {
		t.Log(repo.Name)
	}
}
