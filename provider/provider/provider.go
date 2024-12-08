package provider

type Owner struct {
	Name  string
	IsOrg bool
}

type Repo struct {
	Name        string
	Description string
	AuthToken   string
}

type BackupProvider interface {
	LoadRepos(owner *Owner) ([]string, error)
	MigrateRepo(from *Owner, to *Owner, repo *Repo) (string, error)
	DeleteRepo(owner, repo string) (string, error)
}
