package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
)

type Customer struct {
	ID          persisters.ID `json:"id,omitempty"`
	FirstName   string        `json:"first_name"`
	LastName    string        `json:"last_name"`
	Phone       string        `json:"phone"`
	Email       string        `json:"email"`
	Addresses   []Address     `json:"addresses"`
	Allergies   string        `json:"allergies,omitempty"`
	*Timestamps `json:"timestamps,omitempty"`
}

type CustomerRepository struct {
	Name      string
	localRepo BaseRespository
}

func getCustomerRepository(name string) (CustomerRepository, error) {
	var repo CustomerRepository
	repo.Name = name
	localRepo, err := getRepository(name)
	if err != nil {
		return repo, err
	}

	repo.localRepo = localRepo

	makeCityStateLookup()

	return repo, nil
}

func (repo *CustomerRepository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	return repo.localRepo.Set(id, object)
}

func (repo *CustomerRepository) Get(id persisters.ID, template interface{}) error {
	return repo.localRepo.Get(id, template)
}

func (repo *CustomerRepository) Exists(id persisters.ID) (bool, error) {
	return repo.localRepo.Exists(id)
}

func (repo *CustomerRepository) List() map[persisters.ID]interface{} {
	var customerBuffer Customer
	listing := make(map[persisters.ID]interface{})
	for id := range repo.localRepo.List() {
		repo.Get(id, &customerBuffer)
		listing[id] = customerBuffer.FirstName + " " + customerBuffer.LastName
	}
	return listing
}

func mustMakeCustomerRepo(name string) CustomerRepository {
	repo, err := getCustomerRepository(name)
	if err != nil {
		panic(err)
	}
	return repo
}

var CustomerRepo = mustMakeCustomerRepo("data/customers")
