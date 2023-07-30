package app

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"fmt"
)

func PostCustomer(customer repositories.Customer) {
	fmt.Println("customer object posted to accounting")
}

func PostOrder(order repositories.Order) {
	fmt.Println("order object posted to accouting")
}
