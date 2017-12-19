package secrets

import (
	"github.com/BoxLinker/cicd/models"
)

type builtin struct {
	store models.SecretStore
}

// New returns a new local secret service.
func New(store models.SecretStore) models.SecretService {
	return &builtin{store}
}

func (b *builtin) SecretFind(repo *models.Repo, name string) (*models.Secret, error) {
	return b.store.SecretFind(repo, name)
}

func (b *builtin) SecretList(repo *models.Repo) ([]*models.Secret, error) {
	return b.store.SecretList(repo)
}

func (b *builtin) SecretListBuild(repo *models.Repo, build *models.Build) ([]*models.Secret, error) {
	return b.store.SecretList(repo)
}

func (b *builtin) SecretCreate(repo *models.Repo, in *models.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) SecretUpdate(repo *models.Repo, in *models.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) SecretDelete(repo *models.Repo, name string) error {
	secret, err := b.store.SecretFind(repo, name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}
