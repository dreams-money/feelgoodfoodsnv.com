package app

import (
	persist "DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	repos "DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"errors"
	"log"
	"strconv"
	"strings"
)

type SlotOrderManager struct {
	slots     []repos.FulfillmentSlot
	orders    repos.OrderRepository
	slotFills map[persist.ID]int // Slot => fills
}

type OrderSubmission struct {
	PaymentToken string      `json:"token"`
	Location     string      `json:"location"`
	Order        repos.Order `json:"order"`
}

var OrderManager SlotOrderManager

func LoadOrderManager() error {
	OrderManager = mustMakeManager()
	return nil
}

func (m *SlotOrderManager) LoadOrders() {
	log.Println("Loading orders to order manager")

	for id := range repos.OrderRepo.List() {
		var order repos.Order
		err := repos.OrderRepo.Get(id, &order)
		haltOnError(err)
		m.slotFills[order.FulfillmentSlotID]++
	}
}

func (m *SlotOrderManager) AddOrder(s persist.ID, o repos.Order) (persist.ID, error) {
	err := m.checkIfSlotIsFilled(s)
	if err != nil {
		return 0, err
	}

	repos.ClearMenuItemPhotosFromOrderItems(o.Items)
	FillOrderFees(&o)
	o.GrandTotal = o.Total()
	o.Customer.Timestamps = nil

	id, err := m.orders.Set(0, o)
	if err != nil {
		return 0, err
	}

	m.slotFills[s]++

	return id, err
}

func (m *SlotOrderManager) ClearOrders() {
	m.slotFills = make(map[persist.ID]int)
}

func (m *SlotOrderManager) ReviewOrder(order repos.Order) error {
	var systemMenuItem repos.MenuItem
	systemOrderTotal := float32(0)
	for _, orderMenuItem := range order.Items {
		err := repos.MenuItemRepo.Get(orderMenuItem.ID, &systemMenuItem)
		if err != nil {
			return err
		}

		if systemMenuItem.Price != orderMenuItem.Price {
			return errors.New("Item price mismatch")
		}

		systemOrderTotal += orderMenuItem.Price * float32(orderMenuItem.Quantity)
	}

	if order.SubTotal != systemOrderTotal {
		return errors.New("Order total mismatch")
	}

	var orderSlot repos.FulfillmentSlot
	err := repos.FulfillmentSlotRepo.Get(order.FulfillmentSlotID, &orderSlot)
	if err != nil {
		return err
	}

	if len(orderSlot.ZipCodes) > 0 && len(order.Customer.Addresses) > 0 {
		zipCodeTestResult, err := zipCodeTest(orderSlot.ZipCodes, order.Customer.Addresses[0].Postal)
		if err != nil {
			return err
		}
		if !zipCodeTestResult {
			return errors.New("Order not in appropriate zip code")
		}
	}

	return m.checkIfSlotIsFilled(order.FulfillmentSlotID)
}

func makeManager() (SlotOrderManager, error) {
	var manager SlotOrderManager

	currentWeek, err := GetCurrentWeek()
	if err != nil {
		return manager, err
	}

	for _, day := range currentWeek.WeekDays {
		for _, slot := range day.Slots {
			manager.slots = append(manager.slots, slot)
		}
	}

	manager.orders = repos.OrderRepo
	manager.slotFills = make(map[persist.ID]int)
	manager.LoadOrders()

	return manager, nil
}

func mustMakeManager() SlotOrderManager {
	manager, err := makeManager()
	haltOnError(err)
	return manager
}

func (m *SlotOrderManager) checkIfSlotIsFilled(s persist.ID) error {
	var slot repos.FulfillmentSlot
	err := repos.FulfillmentSlotRepo.Get(s, &slot)
	if err != nil {
		log.Println(err)
		msg := "Slot not found, slot id: " + strconv.Itoa(int(s))
		return errors.New(msg)
	}

	if slot.MaxFils == 0 {
		return nil
	}

	if m.slotFills[s] >= slot.MaxFils {
		return errors.New("Slot filled")
	}

	return nil
}

func FillOrderFees(order *repos.Order) error {
	var orderSlot repos.FulfillmentSlot
	err := repos.FulfillmentSlotRepo.Get(order.FulfillmentSlotID, &orderSlot)
	if err != nil {
		return err
	}

	if order.Fees == nil {
		order.Fees = make(map[string]float32)
	}

	if orderSlot.Fee != nil {
		order.Fees[orderSlot.Fee.Name] = orderSlot.Fee.Amount
	}

	return nil
}

func zipCodeTest(acceptableZipCodes []int, customerZipCode string) (bool, error) {
	customerZipParts := strings.Split(customerZipCode, "-")
	customerZipCode = customerZipParts[0]
	customerZip, err := strconv.Atoi(customerZipCode)
	for _, acceptSlotZip := range acceptableZipCodes {

		if err != nil {
			return false, err
		}
		if acceptSlotZip == customerZip {
			return true, nil
		}
	}

	return false, nil
}

func haltOnError(err error) {
	if err != nil {
		panic(err)
	}
}
