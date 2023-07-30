package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"encoding/json"
	"io/ioutil"
	"os"
)

type NutritionLabel struct {
	Calories      int `json:"calories"`
	Carbohydrates int `json:"carbs"`
	Protiens      int `json:"protiens"`
	Fats          int `json:"fats"`
}

type MenuItem struct {
	ID              persisters.ID `json:"id,omitempty"`
	Name            string        `json:"name"`
	Price           float32       `json:"price,omitempty"`
	Category        string        `json:"category,omitempty"`
	Description     string        `json:"description"`
	Image           string        `json:"base64_image,omitempty"`
	*NutritionLabel `json:"nutrition,omitempty"`
	Qualities       []string `json:"qualities,omitempty"`
	*Timestamps     `json:"timestamps,omitempty"`
}

type Menu map[persisters.ID]MenuItem

type MenuItemRepository struct {
	Name      string
	localRepo BaseRespository
}

func getMenuItemRepository(name string) (MenuItemRepository, error) {
	var repo MenuItemRepository
	repo.Name = name
	localRepo, err := getRepository(name)
	if err != nil {
		return repo, err
	}

	repo.localRepo = localRepo

	return repo, nil
}

func (repo *MenuItemRepository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	return repo.localRepo.Set(id, object)
}

func (repo *MenuItemRepository) Get(id persisters.ID, template interface{}) error {
	return repo.localRepo.Get(id, template)
}

func (repo *MenuItemRepository) Exists(id persisters.ID) (bool, error) {
	return repo.localRepo.Exists(id)
}

func (repo *MenuItemRepository) List() map[persisters.ID]interface{} {
	var menuItem MenuItem
	listing := make(map[persisters.ID]interface{})
	for id := range repo.localRepo.List() {
		repo.Get(id, &menuItem)
		listing[id] = menuItem.Name
	}
	return listing
}

var activeFile = "data/active-menu.json"

func (repo *MenuItemRepository) SetActive(ids []persisters.ID) {
	file, err := json.MarshalIndent(ids, "", "")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(activeFile, file, 644)
	if err != nil {
		panic(err)
	}
}

func (repo *MenuItemRepository) getActiveList() []persisters.ID {
	var active []persisters.ID

	_, err := os.Stat(activeFile)
	if os.IsNotExist(err) {
		return active
	}

	jsonFile, err := ioutil.ReadFile(activeFile)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(jsonFile, &active)

	return active
}

func (repo *MenuItemRepository) GetActive() map[persisters.ID]MenuItem {
	activeList := repo.getActiveList()
	activeListWithDetail := make(map[persisters.ID]MenuItem)

	for _, id := range activeList {
		var menuItem MenuItem
		repo.Get(id, &menuItem)
		activeListWithDetail[id] = menuItem
	}

	return activeListWithDetail
}

type MenuItemListItemDetail struct {
	ID     persisters.ID
	Name   string
	Active bool
}

func (repo *MenuItemRepository) ListWithDetails() map[persisters.ID]MenuItemListItemDetail {
	fullList := repo.List()
	activeList := repo.getActiveList()
	detailedList := make(map[persisters.ID]MenuItemListItemDetail)

	for cursorId, name := range fullList {
		isActive := false
		for _, activeID := range activeList {
			if activeID == cursorId {
				isActive = true
				break
			}
		}
		detailedList[cursorId] = MenuItemListItemDetail{
			ID:     cursorId,
			Name:   name.(string),
			Active: isActive,
		}

	}

	return detailedList
}

func mustMakeMenuItemRepository(name string) MenuItemRepository {
	repo, err := getMenuItemRepository(name)
	if err != nil {
		panic(err)
	}
	return repo
}

var MenuItemRepo = mustMakeMenuItemRepository("data/menuItem")
