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
	Debug   bool   `json:"debug"`
}

type FileBackup struct {
	conf *FileBackupConfig
	his  FileHistory
}

func NewFileBackup(conf *FileBackupConfig) *FileBackup {
	return &FileBackup{conf: conf}
}

func (f *FileBackup) loadHistory() (FileHistory, error) {
	if f.his != nil {
		return f.his, nil
	}
	if _, err := os.Stat(f.conf.History); os.IsNotExist(err) {
		f.his = make(FileHistory)
		return f.his, nil
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
	f.his = repos
	return f.his, nil
}

func (f *FileBackup) syncHistory() error {
	file, err := json.MarshalIndent(f.his, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(f.conf.History, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileBackup) loadRepoHistory(owner, repo string) (*RepoFile, error) {
	repos, err := f.loadHistory()
	if err != nil {
		return nil, err
	}
	var r RepoFile
	if _, ok := repos[owner]; !ok {
		repos[owner] = make(map[string]RepoFile)
	}
	if _, ok := repos[owner][repo]; !ok {
		r = RepoFile{}
	} else {
		r = repos[owner][repo]
	}
	return &r, nil
}

func (f *FileBackup) updateRepoHistory(owner, repo string, file RepoFile) error {
	repos, err := f.loadHistory()
	if err != nil {
		return err
	}
	if _, ok := repos[owner]; !ok {
		repos[owner] = make(map[string]RepoFile)
	}
	repos[owner][repo] = file
	err = f.syncHistory()
	if err != nil {
		return err
	}
	return nil
}

func (f *FileBackup) deleteRepoHistory(owner, repo string) error {
	repos, err := f.loadHistory()
	if err != nil {
		return err
	}
	delete(repos[owner], repo)
	err = f.syncHistory()
	if err != nil {
		return err
	}
	return nil
}

func (f *FileBackup) LoadRepos(owner string) ([]string, error) {
	repos, err := f.loadHistory()
	if err != nil {
		return nil, err
	}
	uRepos, ok := repos[owner]
	if !ok {
		return nil, nil
	}
	var res []string
	for k, _ := range uRepos {
		res = append(res, k)
	}
	return res, nil
}

func (f *FileBackup) MigrateRepo(owner, repoOwner string, repoName, repoDesc string, githubToken string) (string, error) {

	// load history
	r, err := f.loadRepoHistory(owner, repoName)
	if err != nil {
		return "", err
	}

	// create dir if not exist
	dirPath := path.Join(f.conf.Dir, repoOwner)
	if _, e := os.Stat(f.conf.Dir); os.IsNotExist(e) {
		if me := os.MkdirAll(f.conf.Dir, 0755); me != nil {
			return "", me
		}
	}
	if _, e := os.Stat(dirPath); os.IsNotExist(e) {
		if me := os.MkdirAll(dirPath, 0755); me != nil {
			return "", me
		}
	}

	// get repo info from GitHub
	url := fmt.Sprintf("repos/%s/%s", owner, repoName)
	var info struct {
		DefaultBranch string `json:"default_branch"`
		UpdateAt      string `json:"updated_at"`
	}
	err = GithubRequestJson("GET", url, githubToken, &info)
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
	url = fmt.Sprintf("repos/%s/%s/tarball/%s", owner, repoName, info.DefaultBranch)
	if f.conf.Debug {
		tarFile, wErr := os.Create(filePath)
		if wErr != nil {
			return "", wErr
		}
		defer tarFile.Close()
		// write url to file
		_, wErr = tarFile.WriteString(url)
		if wErr != nil {
			return "", wErr
		}
	} else {
		dErr := GithubRequest("GET", url, githubToken, func(resp *http.Response) error {
			tarFile, wErr := os.Create(filePath)
			if wErr != nil {
				return wErr
			}
			defer tarFile.Close()
			_, wErr = io.Copy(tarFile, resp.Body)
			if wErr != nil {
				return wErr
			}
			return nil
		})
		if dErr != nil {
			return "", dErr
		}
		return fmt.Sprintf("file %s downloaded", filePath), nil
	}

	// update history
	r.File = filePath
	r.UpdateAt = info.UpdateAt
	r.Description = repoDesc
	err = f.updateRepoHistory(owner, repoName, *r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file %s updated", filePath), nil
}

func (f *FileBackup) DeleteRepo(owner, repo string) (string, error) {
	r, err := f.loadRepoHistory(owner, repo)
	if err != nil {
		return "", err
	}
	err = os.RemoveAll(r.File)
	if err != nil {
		return "", err
	}
	err = f.deleteRepoHistory(owner, repo)
	if err != nil {
		return "", err
	}
	return "success", nil
}
