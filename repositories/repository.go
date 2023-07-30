package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
)

type Respository interface {
	Set(id persisters.ID, object interface{}) (persisters.ID, error)
	Get(id persisters.ID, template interface{}) error
	Exists(id persisters.ID) (bool, error)
	List() map[persisters.ID]interface{}
}

type BaseRespository struct {
	Name      string
	persister persisters.Persister
}

func getRepository(name string) (BaseRespository, error) {
	var repo BaseRespository
	repo.Name = name
	filePersister, err := persisters.MakeFilePersister(name)
	if err != nil {
		return repo, err
	}

	repo.persister = &filePersister

	return repo, nil
}

func (repo *BaseRespository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	exists, err := repo.Exists(id)
	if err != nil {
		return 0, nil
	}
	if exists {
		return id, repo.persister.Update(id, object)
	}
	return repo.persister.Create(object)
}

func (repo *BaseRespository) Get(id persisters.ID, template interface{}) error {
	return repo.persister.Retrieve(id, &template)
}

func (repo *BaseRespository) Exists(id persisters.ID) (bool, error) {
	if id == 0 {
		return false, nil
	}

	// In memory index?? - FAST
	// I'd need to expose the persister's index
	err := repo.persister.Retrieve(id, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *BaseRespository) List() map[persisters.ID]interface{} {
	return repo.persister.List()
}
