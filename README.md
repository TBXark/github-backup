# github backup

A simple tool to back up GitHub repository to gitea or other provider.

Legacy javascript version can be found [here](./legacy/README.md)

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


### Configuration

```javascript
{
  // Target configuration, will be used to backup the repository
  "targets": [
    {
      // The user or organization to be backed up
      "owner": "GITHUB_OWNER",
      // The token of the repository to be backed up
      "token": "GITHUB_TOKEN",
      // The backup target owner
      "repo_owner": "BACKUP_TARGET_REPO_OWNER",
      "backup": {
        // The backup target type, currently only supports gitea and file
        "type": "file",
        // File configuration, only used when the backup target is file
        "config": {
          // The directory where the backup file is stored
          "dir": "SAVE_DIR",
          // The history json file path
          "history": "FILE_HISTORY_JSON_PATH",
          // When debug is true, files will not be downloaded
          "debug": false
        }
      },
      // Filter rules
      "filter": {
        // When the repository is not matched, the action to be taken, currently only supports delete and ignore
        "unmatched_repo_action": "ignore",
        // Allow rules, only repositories that match the rules will be backed up
        // The rule is a regular expression, the format is :owner/:repo/:private/:fork/:archived
        // For example, the rule [^/]+/[^/]+/0/././. means that only public repositories will be backed up
        "allow_rule": [],
        // Deny rules, repositories that match the rules will not be backed up
        "deny_rule": []
      }
    }
  ],
  // Default configuration, will be used if the target configuration is not sets
  "default_conf": {
    "github_token": "YOUR_GITHUB_TOKEN",
    "repo_owner": "BACKUP_TARGET_REPO_OWNER",
    "backup": {
      "type": "gitea",
      // Gitea configuration, only used when the backup target is gitea
      "config": {
        // Gitea host, You can use your own gitea server
        "host": "GITEA_HOST",
        // Gitea token, You can create a new token in the gitea settings
        "token": "GITEA_TOKEN",
        // Gitea username, You can use your own gitea username
        "auth_username": "GITEA_USERNAME"
      }
    },
    "filter": {
      "unmatched_repo_action": "delete",
      "allow_rule": ["[^/]+/[^/]+/0/././."],
      "deny_rule": ["[^/]+/[^/]+/1/././."]
    }
  }
}
```

### License

**github-backup** is released under the MIT license. See [LICENSE](LICENSE) for details.
