package registry

import (
	"github.com/BoxLinker/cicd/models"
)

type builtin struct {
	store models.RegistryStore
}

// New returns a new local registry service.
func New(store models.RegistryStore) models.RegistryService {
	return &builtin{store}
}

func (b *builtin) RegistryFind(repo *models.Repo, name string) (*models.Registry, error) {
	return b.store.RegistryFind(repo, name)
}

func (b *builtin) RegistryList(repo *models.Repo) ([]*models.Registry, error) {
	return b.store.RegistryList(repo)
}

func (b *builtin) RegistryCreate(repo *models.Repo, in *models.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *builtin) RegistryUpdate(repo *models.Repo, in *models.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *builtin) RegistryDelete(repo *models.Repo, addr string) error {
	registry, err := b.RegistryFind(repo, addr)
	if err != nil {
		return err
	}
	return b.store.RegistryDelete(registry)
}
