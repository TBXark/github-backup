package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type RepoFile struct {
	File        string `json:"file"`
	Description string `json:"description"`
	UpdateAt    string `json:"update_at"`
}

type FileHistory map[string]map[string]RepoFile

type FileBackupConfig struct {
	Dir     string `json:"dir"`
	History string `json:"history"`
}

type FileBackup struct {
	conf *FileBackupConfig
}

func NewFileBackup(conf *FileBackupConfig) *FileBackup {
	return &FileBackup{conf: conf}
}

func (f *FileBackup) LoadHistory() (FileHistory, error) {
	if _, err := os.Stat(f.conf.History); os.IsNotExist(err) {
		return make(FileHistory), nil
	}
	file, err := os.ReadFile(f.conf.History)
	if err != nil {
		return nil, err
	}
	var repos FileHistory
	err = json.Unmarshal(file, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (f *FileBackup) LoadRepos(owner string) ([]string, error) {
	repos, err := f.LoadHistory()
	if err != nil {
		return nil, err
	}
	var res []string
	for k, _ := range repos[owner] {
		res = append(res, k)
	}
	return res, nil
}

func (f *FileBackup) MigrateRepo(owner, repoOwner string, repoName, repoDesc string, githubToken string) (string, error) {

	// load history
	repos, err := f.LoadHistory()
	if err != nil {
		return "", err
	}
	var r RepoFile
	if _, ok := repos[owner]; !ok {
		repos[owner] = make(map[string]RepoFile)
	}
	if _, ok := repos[owner][repoName]; !ok {
		r = RepoFile{}
	} else {
		r = repos[owner][repoName]
	}

	// get repo info from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	req.Header.Add("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	var info struct {
		DefaultBranch string `json:"default_branch"`
		UpdateAt      string `json:"update_at"`
	}
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return "", err
	}

	// ignore if no change
	if r.UpdateAt == info.UpdateAt {
		return "no change", nil
	}

	// remove old file
	filePath := path.Join(f.conf.Dir, repoOwner, fmt.Sprintf("%s.tar.gz", repoName))
	if _, e := os.Stat(filePath); e == nil {
		if re := os.Remove(filePath); re != nil {
			return "", re
		}
	}

	// download file
	url = fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repoName, info.DefaultBranch)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	req.Header.Add("Accept", "application/vnd.github+json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tarFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer tarFile.Close()
	_, err = io.Copy(tarFile, resp.Body)
	if err != nil {
		return "", err
	}

	// update history
	r.File = filePath
	r.UpdateAt = info.UpdateAt
	r.Description = repoDesc
	repos[owner][repoName] = r
	reposRaw, err := json.Marshal(repos)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(f.conf.History, reposRaw, 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file %s updated", filePath), nil
}

func (f *FileBackup) DeleteRepo(owner, repo string) (string, error) {
	repos, err := f.LoadHistory()
	if err != nil {
		return "", err
	}
	r := repos[owner][repo]
	delete(repos[owner], repo)
	file, err := json.Marshal(repos)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(f.conf.History, file, 0644)
	if err != nil {
		return "", err
	}
	err = os.RemoveAll(r.File)
	if err != nil {
		return "", err
	}
	return "success", nil
}
