package repositories

import (
	"os"
	"testing"
)

func getTestCustomer() Customer {
	var addresses []Address
	addresses = append(addresses, Address{
		Type:       "billing",
		Line1:      "170 Koontz Lane #54",
		City:       "Carson City",
		State:      "NV",
		Postal:     "89701",
		Timestamps: Now(),
	})
	addresses = append(addresses, Address{
		Type:       "delivery",
		Line1:      "2201 W. College Parkway",
		City:       "Carson City",
		State:      "NV",
		Postal:     "89706",
		Timestamps: Now(),
	})
	return Customer{
		FirstName:  "Cesar",
		LastName:   "Vega",
		Addresses:  addresses,
		Timestamps: Now(),
	}
}

func getTestMenuItem() MenuItem {
	return MenuItem{
		Name:        "Good Food",
		Description: "You know",
		Category:    "Breakfast",
		Qualities:   []string{"Spicy"},
		Timestamps:  Now(),
	}
}

func getTestFullfillmentSlot() FulfillmentSlot {
	return FulfillmentSlot{
		SlotDescription: "11pm-1pm",
		MaxFils:         20,
		Timestamps:      Now(),
	}
}

func TestWriteCustomer(t *testing.T) {
	repo := CustomerRepo
	customer := getTestCustomer()
	id, err := repo.Set(0, customer)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := repo.Exists(id)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Not written!")
	}
}

func TestReadCustomer(t *testing.T) {
	repo := CustomerRepo
	var storedCustomer Customer
	err := repo.Get(1, &storedCustomer)
	if err != nil {
		t.Fatal(err)
	}
	testCustomer := getTestCustomer()

	if testCustomer.LastName != storedCustomer.LastName {
		t.Fatal("Item read inaccurate")
	}
}

func TestRepoSwitchIsEasy(t *testing.T) {
	repo, err := getRepository("menuItem")
	if err != nil {
		t.Fatal(err)
	}
	menuItem := getTestMenuItem()
	id, err := repo.Set(0, menuItem)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatal("Index should have reset")
	}
	exists, err := repo.Exists(id)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Not written!")
	}
}

func TestFullfillmentSlotReadAndWrite(t *testing.T) {
	repo := FulfillmentSlotRepo

	testSlot := getTestFullfillmentSlot()

	id, err := repo.Set(0, testSlot)
	if err != nil {
		t.Fatal(err)
	}

	var writtenSlot FulfillmentSlot
	repo.Get(id, &writtenSlot)

	m := ""
	if writtenSlot.SlotDescription != testSlot.SlotDescription {
		m = "Slot not written appropirately! Wrote: %v (%v), Local: %v"
		t.Fatalf(m, writtenSlot.SlotDescription, id, testSlot.SlotDescription)
	}
	// if writtenSlot.Slot != id {
	// 	m = "Slot ID not written to file: Wrote: %v, ID: %v"
	// 	t.Fatalf(m, writtenSlot.Slot, id)
	// }
}

func TestCanDisplayListing(t *testing.T) {
	repo, err := getRepository("menuItem")
	if err != nil {
		t.Fatal(err)
	}

	if len(repo.List()) < 1 {
		t.Fatal("Failed to produce listing")
	}

	cleanTests(t)
}

func cleanTests(T *testing.T) {
	err := os.RemoveAll("customer")
	if err != nil {
		T.Fatal(err)
	}
	err = os.RemoveAll("menuItem")
	if err != nil {
		T.Fatal(err)
	}
	err = os.RemoveAll("slots")
	if err != nil {
		T.Fatal(err)
	}
}
