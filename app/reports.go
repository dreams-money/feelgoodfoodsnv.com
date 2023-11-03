package app

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"log"
)

type OrderSheet struct {
	TotalOrders  int
	TotalsByMeal map[string]int // Meal Description => totals
	Orders       map[int]repositories.Order
	TotalRevenue float32
}

func CreateOrderSheet() OrderSheet {
	var orderSheet OrderSheet
	orderSheet.TotalsByMeal = make(map[string]int)
	orderSheet.Orders = make(map[int]repositories.Order)

	for orderID := range repositories.OrderRepo.List() {
		var order repositories.Order
		repositories.OrderRepo.Get(orderID, &order)

		orderSheet.TotalOrders++
		for _, orderItem := range order.Items {
			if orderItem.MenuItem.Name != "" {
				orderSheet.TotalsByMeal[orderItem.MenuItem.Name] += orderItem.Quantity
			}
		}

		orderSheet.Orders[int(orderID)] = order
		orderSheet.TotalRevenue += order.GrandTotal
	}

	return orderSheet
}

type DeliveryReport struct {
	TotalOrders int
	Areas       DeliverySheet
}

type DeliveryArea struct {
	TotalOrders int
	OrderMap    OrderMap
}

type OrderMap map[persisters.ID]repositories.Order
type DeliverySheet map[string]DeliveryArea // Area => orders

func CreateDeliverySheet() DeliveryReport { //At this level "fulfillment" and delivery are the same.
	var report DeliveryReport

	report.Areas = make(DeliverySheet)

	for orderID := range repositories.OrderRepo.List() {
		var order repositories.Order
		err := repositories.OrderRepo.Get(orderID, &order)
		if err != nil {
			log.Println(err)
		}

		var slot repositories.FulfillmentSlot
		err = repositories.FulfillmentSlotRepo.Get(order.FulfillmentSlotID, &slot)
		if err != nil {
			log.Println(err)
		}
		order.FulfillmentSlot = slot

		var city string
		if len(order.Customer.Addresses) > 0 {
			city = order.Customer.Addresses[0].City
		} else {
			city = "No Address"
		}

		ordersForCity, found := report.Areas[city]
		if !found {
			ordersForCity = DeliveryArea{}
		}
		ordersForCity.TotalOrders++
		if ordersForCity.OrderMap == nil {
			ordersForCity.OrderMap = make(OrderMap)
		}
		ordersForCity.OrderMap[orderID] = order
		report.Areas[city] = ordersForCity

		report.TotalOrders++
	}

	return report
}
