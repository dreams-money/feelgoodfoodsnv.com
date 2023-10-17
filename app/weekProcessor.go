package app

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cp "github.com/otiai10/copy"
)

const cutOffTime = "Wed 12:01AM"

var CurrentWeekID persisters.ID

func RunCutoffSchedule(syncer chan struct{}) {
	nextCutOffTime, err := createUpcomingCutoffTime(cutOffTime)
	panicOnError(err)

	CurrentWeekID = getLastestWeekID()

	log.Println("Week processor starting, on week: " + strconv.Itoa(int(CurrentWeekID)))

	go func() {
		for {
			log.Println("Next cutoff time: " + nextCutOffTime.Format("Mon Jan _2 03:04PM MST 2006"))

			time.Sleep(time.Until(nextCutOffTime))

			log.Println("Storing last week's orders")
			StoreCurrentWeeksOrders()
			log.Println("Generating new week")
			CreateNewWeek()
			OrderManager.ClearOrders()

			nextCutOffTime = nextCutOffTime.AddDate(0, 0, 7)

			syncer <- struct{}{}
		}
	}()
}

func GetCurrentWeek() (repositories.Week, error) {
	var currentWeek repositories.Week
	err := repositories.WeekRepo.Get(CurrentWeekID, &currentWeek)
	if err != nil {
		return currentWeek, err
	}

	for dayKey, day := range currentWeek.WeekDays {
		for slotID := range day.Slots {
			err := OrderManager.checkIfSlotIsFilled(slotID)
			if err != nil {
				delete(currentWeek.WeekDays[dayKey].Slots, slotID)
			}
		}
	}

	return currentWeek, nil
}

func CreateNewWeek() {
	id, err := repositories.WeekRepo.Set(
		persisters.ID(0), repositories.CreateNewWeek())
	panicOnError(err)
	CurrentWeekID = id
}

func StoreCurrentWeeksOrders() {
	var week repositories.Week
	var order repositories.Order
	var slot repositories.FulfillmentSlot

	err := repositories.WeekRepo.Get(CurrentWeekID, &week)
	panicOnError(err)

	for orderID := range repositories.OrderRepo.List() {

		err = repositories.OrderRepo.Get(orderID, &order)
		logOnError(err)

		slotID := order.FulfillmentSlotID
		err = repositories.FulfillmentSlotRepo.Get(slotID, &slot)
		logOnError(err)

		dayOfWeekString := repositories.DayMap[slot.DayOfWeek]

		weekSlot := week.WeekDays[dayOfWeekString].Slots[slotID]
		weekSlot.Orders = append(weekSlot.Orders, order)
		week.WeekDays[dayOfWeekString].Slots[slotID] = weekSlot
	}

	_, err = repositories.WeekRepo.Set(CurrentWeekID, week)
	panicOnError(err)
}

func cacheThisWeeksOrders(week repositories.Week) error {
	weekDescription := strings.Replace(week.Description, " ", "-", -1)

	currentDir := "./data/orders"
	cachedDir := "./data/orders-" + weekDescription
	err := cp.Copy(currentDir, cachedDir)
	if err != nil {
		return err
	}

	currentDirSize, err := dirSize(currentDir)
	if err != nil {
		return err
	}
	cachedDirSize, err := dirSize(cachedDir)
	if err != nil {
		return err
	}

	if currentDirSize != cachedDirSize {
		return errors.New("Failed to cache week - dir sizes don't match! " +
			strconv.Itoa(int(currentDirSize)) + " " + strconv.Itoa(int(cachedDirSize)))
	}

	return removeDirContents(currentDir)
}

func createUpcomingCutoffTime(configOption string) (time.Time, error) {
	dayToCutOff, hourToCutOff, minuteToCutOff, err := parseCutOffTimeConfig(configOption)
	if err != nil {
		return time.Now(), err
	}

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Now(), err
	}

	now := time.Now().In(location)

	daysToCutOffDay := dayToCutOff - int(now.Weekday())
	if int(now.Weekday()) >= dayToCutOff {
		daysToCutOffDay += 7
	}
	nextCutOffTime := now.AddDate(0, 0, daysToCutOffDay)

	nextCutOffTime = time.Date(nextCutOffTime.Year(),
		nextCutOffTime.Month(), nextCutOffTime.Day(),
		hourToCutOff, minuteToCutOff, 0, 0,
		nextCutOffTime.Location())

	return nextCutOffTime, nil
}

func parseCutOffTimeConfig(time string) (int, int, int, error) {
	parts := strings.Split(time, " ")
	dayToCutOff, err := repositories.WeekDayToKey(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}

	timeToCutOff := parts[1]
	parts = strings.Split(timeToCutOff, ":")

	hourToCutOff, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}
	amOrPm := parts[1][2:4]
	// if amOrPm == "PM" || amOrPm == "pm" {
	// 	hourToCutOff += 12
	// } else
	if amOrPm == "AM" || amOrPm == "am" {
		hourToCutOff -= 12
	}

	minuteToCutOff, err := strconv.Atoi(parts[1][0:2])
	if err != nil {
		return 0, 0, 0, err
	}

	return dayToCutOff, hourToCutOff, minuteToCutOff, nil
}

func getLastestWeekID() persisters.ID {
	maxID := persisters.ID(0)
	for weekId := range repositories.WeekRepo.List() {
		if int(weekId) > int(maxID) {
			maxID = persisters.ID(weekId)
		}
	}
	return maxID
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func logOnError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func removeDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
