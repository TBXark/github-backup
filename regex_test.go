package main

import "testing"

func TestIsMatchRepoDescription(t *testing.T) {
	if !IsMatchRepoIdentity("tbxark/backup/1/0/0", "[^/]+/[^/]+/1/0/0") {
		t.Fatal("not match")
	}
	if IsMatchRepoIdentity("tbxark/backup/1/0/0", "[^/]+/[^/]+/0/0/0") {
		t.Fatal("match")
	}
	if !IsMatchRepoIdentity("tbxark/backup/1/0/0", "[^/]+/[^/]+/1/./.") {
		t.Fatal("match")
	}
	if IsMatchRepoIdentity("tbxark/backup/1/0/0", "[^/]+/[^/]+/0/././.") {
		t.Fatal("match")
	}
}
