package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
)

type OrderItem struct {
	MenuItem `json:"menu_item"`
	Quantity int     `json:"quantity"`
	Price    float32 `json:"price"`
}

type Order struct {
	ID                persisters.ID      `json:"id,omitempty"`
	FulfillmentSlotID persisters.ID      `json:"slot_id"`
	Customer          Customer           `json:"customer"`
	Items             []OrderItem        `json:"items"`
	SubTotal          float32            `json:"sub_total"`
	Fees              map[string]float32 `json:"fees,omitempty"`
	GrandTotal        float32            `json:"total,omitempty"`
	*Timestamps       `json:"timestamps,omitempty"`
	FulfillmentSlot
}

func (order *Order) Total() float32 {
	feeTotal := float32(0)
	for _, fee := range order.Fees {
		feeTotal += fee
	}
	return order.SubTotal + feeTotal
}

type OrderRepository struct {
	Name      string
	localRepo BaseRespository
}

func FillMenuItemDetails(order *Order) error {
	var orderItems []OrderItem
	for _, item := range *&order.Items {
		var systemMenuItem MenuItem
		err := MenuItemRepo.Get(item.MenuItem.ID, &systemMenuItem)
		if err != nil {
			return err
		}
		orderItems = append(orderItems, OrderItem{
			Quantity: item.Quantity,
			MenuItem: MenuItem{
				Name:        systemMenuItem.Name,
				Image:       systemMenuItem.Image,
				Description: systemMenuItem.Description,
			},
			Price: item.Price,
		})
	}

	order.Items = orderItems

	return nil
}

func getOrderRepository(name string) (OrderRepository, error) {
	var repo OrderRepository
	repo.Name = name
	localRepo, err := getRepository(name)
	if err != nil {
		return repo, err
	}

	repo.localRepo = localRepo

	return repo, nil
}

func (repo *OrderRepository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	return repo.localRepo.Set(id, object)
}

func (repo *OrderRepository) Get(id persisters.ID, template interface{}) error {
	return repo.localRepo.Get(id, template)
}

func (repo *OrderRepository) Exists(id persisters.ID) (bool, error) {
	return repo.localRepo.Exists(id)
}

func (repo *OrderRepository) List() map[persisters.ID]interface{} {
	return repo.localRepo.List()
}

func mustMakeOrderRepo(name string) OrderRepository {
	repo, err := getOrderRepository(name)
	if err != nil {
		panic(err)
	}
	return repo
}

var OrderRepo = mustMakeOrderRepo("data/orders")

func ClearMenuItemPhotosFromOrderItems(items []OrderItem) {
	for i := range items {
		items[i].MenuItem.Image = ""
	}
}
