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
  "host": "",
  "gitea_token": "",
  "auth_username": "",
  "default_conf": {
    "github_token": "",
    "repo_owner": "",
    "backup": {
      "type": "",
      "config": {
        "host": "a",
        "token": "b",
        "auth_username": "c"
      }
    }
  },
  "targets": [
    {
      "owner": "",
      "token": "",
      "repo_owner": "",
      "backup": null
    }
  ]
}


```