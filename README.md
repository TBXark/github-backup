# github backup

A simple tool to backup github repository to gitea or other provider.

### Installation

```bash
go install github.com/TBXark/github-backup@latest
````

### Usage
```
Usage of github-backup:
  -config string
        config file (default "config.json")
  -help
        show help
  -version
        show version

```

```bash
github-backup --config config.json
```


### Config

```json
{
  "default_conf": {
    "github_token": "YOUR_GITHUB_TOKEN",
    "repo_owner": "BACKUP_TARGET_REPO_OWNER",
    "backup": {
      "type": "gitea",
      "config": {
        "host": "GITEA_HOST",
        "token": "GITEA_TOKEN",
        "auth_username": "GITEA_USERNAME"
      }
    },
    "filter": {
      "unmatched_repo_action": "delete",
      "allow_rule": [
        "[^/]+/[^/]+/0/././."
      ],
      "deny_rule": [
        "[^/]+/[^/]+/1/././."
      ]
    }
  },
  "targets": [
    {
      "owner": "GITHUB_OWNER",
      "token": "GITHUB_TOKEN",
      "repo_owner": "BACKUP_TARGET_REPO_OWNER",
      "backup": {
        "type": "file",
        "config": {
          "dir": "SAVE_DIR",
          "history": "FILE_HISTORY_JSON_PATH",
          "debug": false
        }
      },
      "filter": {
        "unmatched_repo_action": "ignore",
        "allow_rule": null,
        "deny_rule": null
      }
    }
  ]
}
```