package app

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	repos "DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"log"
)

var serverConfiguration config.Config

func ProcessOrders(cfg config.Config, syncer chan struct{}) {
	serverConfiguration = cfg
	initialID := persisters.ID(1)
	weekDataExists, _ := repos.WeekRepo.Exists(initialID)
	slotDataExists, _ := repos.FulfillmentSlotRepo.Exists(initialID)
	activeMenu := repos.MenuItemRepo.GetActive()

	if slotDataExists && len(activeMenu) > 0 {
		if !weekDataExists {
			log.Println("Generating Initial Week")
			CreateNewWeek()
		}
		RunCutoffSchedule(syncer)
		LoadOrderManager()
	} else {
		log.Println("NOTICE - configure your menu and order fulfillment slots then restart the " +
			"system to load the order manager.  This required restart is required only one time.")
	}
}
