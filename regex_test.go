package main

import "testing"

func TestIsMatchRepoDescription(t *testing.T) {
	repoRegex := "[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+"
	publicRegex := repoRegex + "/0/[01]/[01]"
	privateRegex := repoRegex + "/1/[01]/[01]"

	t.Log(publicRegex)
	t.Log(privateRegex)
	publicCases := map[string]bool{
		"tbxark/backup/0/0/0":     true,
		"tbxark/backup/0/0/1":     true,
		"tbxark/backup/0/1/0":     true,
		"tbxark/backup/0/1/1":     true,
		"tbxark-arc/backup/0/0/0": true,
		"tbxark-arc/backup/0/0/1": true,
		"tbxark-arc/backup/0/1/0": true,
		"tbxark-arc/backup/0/1/1": true,
	}
	privateCases := map[string]bool{
		"tbxark/backup/1/0/0":         true,
		"tbxark/backup/1/0/1":         true,
		"tbxark/backup/1/1/0":         true,
		"tbxark/backup/1/1/1":         true,
		"tbxark-arc/backup-arc/1/0/0": true,
		"tbxark-arc/backup-arc/1/0/1": true,
		"tbxark-arc/backup-arc/1/1/0": true,
		"tbxark-arc/backup-arc/1/1/1": true,
	}
	for c, v := range publicCases {
		if !IsMatchRepoIdentity(c, publicRegex) {
			t.Errorf("public case %s expect %v but %v", c, v, false)
		}
	}
	for c, v := range privateCases {
		if !IsMatchRepoIdentity(c, privateRegex) {
			t.Errorf("private case %s expect %v but %v", c, v, false)
		}
	}

}
