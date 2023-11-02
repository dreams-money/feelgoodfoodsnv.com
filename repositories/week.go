package repositories

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"errors"
	"strconv"
	"time"
)

var DayMap = map[int]string{
	1: "Mon",
	2: "Tues",
	3: "Wed",
	4: "Thur",
	5: "Fri",
	6: "Sat",
	7: "Sun",
}

type WeekDay struct {
	Description string                            `json:"description"` // Full day string : Monday, Tues, etc
	Slots       map[persisters.ID]FulfillmentSlot `json:"slots"`
}

type Week struct {
	Description string             `json:"description"` // "Week of: <text>"
	WeekDays    map[string]WeekDay `json:"days"`
	Menu        `json:"menu"`
}

func WeekDayToKey(inputDay string) (int, error) {
	for key, day := range DayMap {
		if day == inputDay {
			return key, nil
		}
	}

	return 0, errors.New("Key not found for day: " + inputDay)
}

func CreateNewWeek() Week {
	var week Week
	week.WeekDays = make(map[string]WeekDay)
	daySlots := FulfillmentSlotRepo.List()
	orders := OrderRepo.List()

	now := time.Now()
	// We'll scroll to a relative position to work from
	sundayDateDay := now.Day() - int(now.Weekday())
	sundayDate := time.Date(now.Year(), now.Month(), sundayDateDay,
		0, 0, 0, 0, now.Location())

	// Slots are for next week
	sundayDate = sundayDate.AddDate(0, 0, 7)

	// Monday
	week.Description = formatDate(sundayDate.AddDate(0, 0, 1))

	// Merge slots and orders
	for slotId := range daySlots {
		var slot FulfillmentSlot
		FulfillmentSlotRepo.Get(slotId, &slot)
		slot.ID = slotId

		// Merge orders to slot before adding back to weekday
		for orderId := range orders {
			var order Order
			OrderRepo.Get(orderId, &order)
			if order.FulfillmentSlotID == slot.ID {
				slot.Orders = append(slot.Orders, order)
			}
		}

		weekDayDescription := DayMap[slot.DayOfWeek]

		weekDay, weekDayExists := week.WeekDays[weekDayDescription]

		if !weekDayExists {
			weekDay = WeekDay{
				Description: sundayDate.
					AddDate(0, 0, slot.DayOfWeek).Format("2"),
				Slots: make(map[persisters.ID]FulfillmentSlot),
			}
		}

		weekDay.Slots[slotId] = slot
		week.WeekDays[weekDayDescription] = weekDay
	}

	// Merge the Menu
	week.Menu = MenuItemRepo.GetActive()

	return week
}

func GetCurrentWeekName() string {
	// DRY 1
	now := time.Now()
	sundayDateDay := now.Day() - int(now.Weekday())
	sundayDate := time.Date(now.Year(), now.Month(),
		sundayDateDay, 0, 0, 0, 0, now.Location())
	tillDate := sundayDate.AddDate(0, 0, 7)
	if sundayDate.Month() == tillDate.Month() {
		return sundayDate.Format("January 2") +
			"-" + strconv.Itoa(tillDate.Day())
	} else {
		return sundayDate.Format("January 2") +
			"-" + tillDate.Format("January 2")
	}
}

func GetNextWeekName() string {
	// DRY 1
	now := time.Now()
	sundayDateDay := now.Day() - int(now.Weekday())
	sundayDate := time.Date(now.Year(), now.Month(),
		sundayDateDay, 0, 0, 0, 0, now.Location())
	sundayDate = sundayDate.AddDate(0, 0, 7)
	tillDate := sundayDate.AddDate(0, 0, 7)
	if sundayDate.Month() == tillDate.Month() {
		return sundayDate.Format("January 2") +
			"-" + strconv.Itoa(tillDate.Day())
	} else {
		return sundayDate.Format("January 2") +
			"-" + tillDate.Format("January 2")
	}
}

func (w *Week) GetPickupSlots() map[string]WeekDay {
	// DRY 2
	weekDays := make(map[string]WeekDay)
	pickupSlots := make(map[persisters.ID]FulfillmentSlot)

	for dayKey, day := range w.WeekDays {
		for slotID, slot := range day.Slots {
			if slot.Type == "pickup" {
				pickupSlots[slotID] = slot
			}
		}

		if len(pickupSlots) > 0 {
			weekDays[dayKey] = WeekDay{
				Description: day.Description,
				Slots:       pickupSlots,
			}
			pickupSlots = make(map[persisters.ID]FulfillmentSlot)
		}
	}

	return weekDays
}

func (w *Week) GetDeliverySlots() map[string]WeekDay {
	// DRY 2
	weekDays := make(map[string]WeekDay)
	deliverySlots := make(map[persisters.ID]FulfillmentSlot)

	for dayKey, day := range w.WeekDays {
		for slotID, slot := range day.Slots {
			if slot.Type == "delivery" {
				deliverySlots[slotID] = slot
			}
		}

		if len(deliverySlots) > 0 {
			weekDays[dayKey] = WeekDay{
				Description: day.Description,
				Slots:       deliverySlots,
			}
			deliverySlots = make(map[persisters.ID]FulfillmentSlot)
		}
	}

	return weekDays
}

type WeekRepository struct {
	Name string
	repo BaseRespository
}

func getWeekRepository(name string) (WeekRepository, error) {
	var repo WeekRepository
	repo.Name = name
	localRepo, err := getRepository(name)
	if err != nil {
		return repo, err
	}

	repo.repo = localRepo

	return repo, nil
}

func (repo *WeekRepository) Set(id persisters.ID, object interface{}) (persisters.ID, error) {
	return repo.repo.Set(id, object)
}

func (repo *WeekRepository) Get(id persisters.ID, template interface{}) error {
	return repo.repo.Get(id, template)
}

func (repo *WeekRepository) Exists(id persisters.ID) (bool, error) {
	return repo.repo.Exists(id)
}

func (repo *WeekRepository) List() map[persisters.ID]interface{} {
	return repo.repo.List()
}

func formatDate(t time.Time) string {
	suffix := "th"
	switch t.Day() {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return t.Format("Monday Jan 2" + suffix)
}

func mustMakeWeekRepository(name string) WeekRepository {
	repo, err := getWeekRepository(name)
	if err != nil {
		panic(err)
	}
	return repo
}

var WeekRepo = mustMakeWeekRepository("data/weeks")
