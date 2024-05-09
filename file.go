package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type FileBackupConfig struct {
	Dir     string `json:"dir"`
	History string `json:"history"`
	Debug   bool   `json:"debug"`
}

type FileBackup struct {
	conf    *FileBackupConfig
	history *fileHistory
}

func NewFileBackup(conf *FileBackupConfig) (*FileBackup, error) {
	his, err := newFileHistory(conf.History)
	if err != nil {
		return nil, err
	}
	return &FileBackup{conf: conf, history: his}, nil
}

func (f *FileBackup) LoadRepos(owner string, isOrg bool) ([]string, error) {
	repos := f.history.loadRepos(owner)
	var res []string
	for k, _ := range repos {
		res = append(res, k)
	}
	return res, nil
}

func (f *FileBackup) MigrateRepo(owner, repoOwner string, isOwnerOrg, isRepoOwnerOrg bool, repoName, repoDesc string, githubToken string) (string, error) {

	// load history
	r := f.history.loadRepoHistory(owner, repoName)

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
	err := GithubRequestJson("GET", url, githubToken, &info)
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
	err = f.history.updateRepoHistory(owner, repoName, *r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file %s updated", filePath), nil
}

func (f *FileBackup) DeleteRepo(owner, repo string) (string, error) {
	r := f.history.loadRepoHistory(owner, repo)
	err := os.RemoveAll(r.File)
	if err != nil {
		return "", err
	}
	err = f.history.deleteRepoHistory(owner, repo)
	if err != nil {
		return "", err
	}
	return "success", nil
}

// FileHistory is a struct to store the history of the file

type repoFile struct {
	File        string `json:"file"`
	Description string `json:"description"`
	UpdateAt    string `json:"update_at"`
}

type fileHistory struct {
	path  string
	store map[string]map[string]repoFile
}

func newFileHistory(p string) (*fileHistory, error) {
	f := &fileHistory{path: p}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		f.store = make(map[string]map[string]repoFile)
		return f, nil
	}
	file, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var repos map[string]map[string]repoFile
	err = json.Unmarshal(file, &repos)
	if err != nil {
		return nil, err
	}
	f.store = repos
	return f, nil
}

func (f *fileHistory) sync() error {
	file, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(f.path, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileHistory) loadRepos(owner string) map[string]repoFile {
	return f.store[owner]
}

func (f *fileHistory) loadRepoHistory(owner, repo string) *repoFile {
	var r repoFile
	if _, ok := f.store[owner]; !ok {
		f.store[owner] = make(map[string]repoFile)
	}
	if _, ok := f.store[owner][repo]; !ok {
		r = repoFile{}
	} else {
		r = f.store[owner][repo]
	}
	return &r
}

func (f *fileHistory) updateRepoHistory(owner, repo string, file repoFile) error {
	if _, ok := f.store[owner]; !ok {
		f.store[owner] = make(map[string]repoFile)
	}
	f.store[owner][repo] = file
	return f.sync()
}

func (f *fileHistory) deleteRepoHistory(owner, repo string) error {
	delete(f.store[owner], repo)
	return f.sync()
}
