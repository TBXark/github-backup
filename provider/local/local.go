package local

import (
	"fmt"
	"github.com/TBXark/github-backup/provider/provider"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Root      string `json:"root"`
	Questions bool   `json:"questions"`
}

var _ provider.Provider = &Local{}

type Local struct {
	conf *Config
}

func NewLocal(conf *Config) *Local {
	return &Local{conf: conf}
}

func (l *Local) LoadRepos(owner *provider.Owner) ([]string, error) {
	ownerPath := filepath.Join(l.conf.Root, owner.Name)
	dirEntries, err := os.ReadDir(ownerPath)
	if err != nil {
		return nil, err
	}
	repos := make([]string, 0)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() && isGitRepository(filepath.Join(ownerPath, dirEntry.Name())) {
			repos = append(repos, dirEntry.Name())
		}
	}
	return repos, nil
}

func (l *Local) MigrateRepo(from *provider.Owner, to *provider.Owner, repo *provider.Repo) (string, error) {
	if l.conf.Questions && !question(fmt.Sprintf("Are you sure you want to migrate %s/%s to %s/%s? [y/n]: ", from.Name, repo.Name, to.Name, repo.Name)) {
		return "skip", nil
	}
	ownerPath := filepath.Join(l.conf.Root, to.Name)
	_, err := os.Stat(ownerPath)
	if err != nil {
		if os.IsNotExist(err) {
			if e := os.MkdirAll(ownerPath, os.ModePerm); e != nil {
				return "", e
			}
		} else {
			return "", err
		}
	}
	repoPath := filepath.Join(ownerPath, repo.Name)
	_, err = os.Stat(repoPath)
	gitUrl := fmt.Sprintf("git@github.com:%s/%s.git", from.Name, repo.Name)
	if err != nil {
		if os.IsNotExist(err) {
			err = gitClone(gitUrl, repoPath)
			if err != nil {
				return "success", err
			}
		} else {
			return "", err
		}
	}
	err = gitPull(repoPath)
	if err != nil {
		return "fail", err
	}
	return "success", nil
}

func (l *Local) DeleteRepo(owner, repo string) (string, error) {
	if l.conf.Questions && !question(fmt.Sprintf("Are you sure you want to delete %s/%s? [y/n]: ", owner, repo)) {
		return "skip", nil
	}
	repoPath := filepath.Join(l.conf.Root, owner, repo)
	err := os.RemoveAll(repoPath)
	if err != nil {
		return "fail", err
	}
	return "success", nil
}

func isGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

func gitClone(url, path string) error {
	log.Printf("cloning %s", url)
	cmd := exec.Command("git", "clone", url, path)
	return cmd.Run()
}

func gitPull(path string) error {
	log.Printf("pulling %s", path)
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	return cmd.Run()
}

func question(message string) bool {
	var response string
	fmt.Print(message)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	return response == "y"
}
