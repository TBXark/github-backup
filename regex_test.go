package main

import "testing"

func TestIsMatchRepoDescription(t *testing.T) {
	if !IsMatchRepoDescription("tbxark/backup/1/0/0", "[^/]+/[^/]+/1/0/0") {
		t.Fatal("not match")
	}
	if IsMatchRepoDescription("tbxark/backup/1/0/0", "[^/]+/[^/]+/0/0/0") {
		t.Fatal("match")
	}
	if !IsMatchRepoDescription("tbxark/backup/1/0/0", "[^/]+/[^/]+/1/./.") {
		t.Fatal("match")
	}
	if IsMatchRepoDescription("tbxark/backup/1/0/0", "[^/]+/[^/]+/0/././.") {
		t.Fatal("match")
	}
}
