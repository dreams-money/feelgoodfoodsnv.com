package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"errors"
)

type FulfillmentSlot struct {
	ID              persisters.ID `json:"slot_id"`
	Type            string        `json:"type"`
	DayOfWeek       int           `json:"day"`
	SlotDescription string        `json:"slot_description"`
	MaxFils         int           `json:"max_fills"`
	Orders          []Order       `json:"orders"`
	*Fee            `json:"fee"`
	ZipCodes        []int `json:"zip_codes"`
	*Timestamps     `json:"timestamps,omitempty"`
}

func (s *FulfillmentSlot) AddOrder(order Order) error {
	if s.IsFilled() {
		return errors.New("Slot is filled")
	}

	s.Orders = append(s.Orders, order)

	return nil
}

func (s *FulfillmentSlot) IsFilled() bool {
	return len(s.Orders) >= s.MaxFils
}

type FulfillmentSlotRepository struct {
	Name      string
	localRepo BaseRespository
}

func getFulfillmentSlotRepository(name string) (FulfillmentSlotRepository, error) {
	var repo FulfillmentSlotRepository
	repo.Name = name
	localRepo, err := getRepository(name)
	if err != nil {
		return repo, err
	}

	repo.localRepo = localRepo

	return repo, nil
}

func (repo *FulfillmentSlotRepository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	return repo.localRepo.Set(id, object)
}

func (repo *FulfillmentSlotRepository) Get(id persisters.ID, template interface{}) error {
	return repo.localRepo.Get(id, template)
}

func (repo *FulfillmentSlotRepository) Exists(id persisters.ID) (bool, error) {
	return repo.localRepo.Exists(id)
}

func (repo *FulfillmentSlotRepository) List() map[persisters.ID]interface{} {
	var slot FulfillmentSlot
	listing := make(map[persisters.ID]interface{})
	for id := range repo.localRepo.List() {
		repo.Get(id, &slot)
		listing[id] = DayMap[slot.DayOfWeek] + " " + slot.SlotDescription
	}
	return listing
}

func mustMakeFulfillmentSlotRepo(name string) FulfillmentSlotRepository {
	repo, err := getFulfillmentSlotRepository(name)
	if err != nil {
		panic(err)
	}
	return repo
}

var FulfillmentSlotRepo = mustMakeFulfillmentSlotRepo("data/slots")
