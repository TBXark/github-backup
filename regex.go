package main

import (
	"log"
	"path"
	"regexp"
)

// RepoIdentity 生产仓库描述用于正则匹配 example: tbxark/backup/1/0/0  :owner/:repo/:private/:fork/:archived
func RepoIdentity(owner, repo string, private, fork, archived bool) string {
	bool2str := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}
	return path.Join(owner, repo, bool2str(private), bool2str(fork), bool2str(archived))
}

func IsMatchRepoIdentity(id string, reg ...string) bool {
	for _, r := range reg {
		regx := regexp.MustCompile(r)
		if regx.MatchString(id) {
			log.Printf("match %s %s", id, r)
			return true
		}
	}
	return false
}
