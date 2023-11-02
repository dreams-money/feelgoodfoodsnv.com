package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/app"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/templates"
)

var serverConfiguration config.Config

func OrderPage(w http.ResponseWriter, r *http.Request) {
	toView := make(map[string]interface{})

	currentWeek, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)
	toView["MenuItems"] = currentWeek.Menu
	toView["WeekOf"] = currentWeek.Description
	toView["IsLocked"] = app.OrderManager.IsLocked()

	executeTemplate("order", w, toView)
}

func DeliveryPage(w http.ResponseWriter, r *http.Request) {
	toView := make(map[string]interface{})

	week, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)
	toView["Description"] = week.Description
	toView["Slots"] = week.GetDeliverySlots()
	toView["IsLocked"] = app.OrderManager.IsLocked()

	executeTemplate("delivery", w, toView)
}

func PickupPage(w http.ResponseWriter, r *http.Request) {
	toView := make(map[string]interface{})

	week, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)

	toView["PickupSlots"] = week.GetPickupSlots()

	if serverConfiguration.Environment == "dev" {
		toView["APIKey"] = serverConfiguration.GoogleMapsAPIKeys.Development
	} else {
		toView["APIKey"] = serverConfiguration.GoogleMapsAPIKeys.Production
	}

	toView["WeekOf"] = week.Description
	toView["IsLocked"] = app.OrderManager.IsLocked()

	executeTemplate("pickup", w, toView)
}

func ReviewPage(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	toView := make(map[string]interface{})
	week, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)
	toView["WeekOf"] = week.Description
	toView["IsLocked"] = app.OrderManager.IsLocked()

	fetchOrderFromURL(urlParams, toView)

	executeTemplate("review", w, toView)
}

func PayPage(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	toView := make(map[string]interface{})

	fetchOrderFromURL(urlParams, toView)

	order := toView["Order"].(repositories.Order)
	var slot repositories.FulfillmentSlot
	err := repositories.FulfillmentSlotRepo.Get(order.FulfillmentSlotID, &slot)
	errorCheckHandleGraceful(err, toView)
	toView["SlotType"] = slot.Type

	toView["IsLocked"] = app.OrderManager.IsLocked()
	week, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)
	toView["WeekOf"] = week.Description

	executeTemplate("pay", w, toView)
}

func SchedulePage(w http.ResponseWriter, r *http.Request) {
	toView := make(map[string]interface{})
	week, err := app.GetCurrentWeek()
	errorCheckHandleGraceful(err, toView)
	toView["WeekOf"] = week.Description
	toView["IsLocked"] = app.OrderManager.IsLocked()
	executeTemplate("schedule", w, toView)
}

func Submit(w http.ResponseWriter, r *http.Request) {
	// This should be part of API..
	if r.Method != http.MethodPost {
		return
	}

	log.Printf("%s: Received payment", r.RemoteAddr)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var orderSubmission app.OrderSubmission
	err = json.Unmarshal(body, &orderSubmission)
	if err != nil {
		log.Println(err)
	}

	order := &orderSubmission.Order

	if order.Customer.Email == "" {
		log.Println("Order must have email")
		return
	}

	err = app.OrderManager.ReviewOrder(*order)
	if err != nil {
		log.Println("Order failed review: " + err.Error())
		return
	}

	if len(order.Customer.Addresses) > 0 {
		err = order.Customer.Addresses[0].FillCityStateFromZip()
		if err != nil {
			log.Println("Failed to fill city state: " + err.Error())
		}
	}

	err = app.FillOrderFees(order)
	if err != nil {
		log.Println(err)
	}

	token := orderSubmission.PaymentToken
	amount := fmt.Sprintf("%.2f", order.Total())

	paymentResponse, err := app.SubmitPayment(token, amount)
	if err != nil {
		log.Println(err)
	}

	newID := 0
	if paymentResponse.Payment.Status == "COMPLETED" {
		log.Printf("%s: Payment success", r.RemoteAddr)

		err = repositories.FillMenuItemDetails(order)
		if err == nil {
			persisterID, err := app.OrderManager.AddOrder(order.FulfillmentSlotID, *order)
			if err == nil {
				log.Printf("%s: Order saved", r.RemoteAddr)
				err = app.SendOrderReceiptEmail(*order, persisterID)

				// Post order to accounting at some point.

				if err != nil {
					log.Println("Failed to send order receipt", err)
				} else {
					log.Printf("%s: Receipt email sent", r.RemoteAddr)
				}
			}
			newID = int(persisterID)
		}

		if err != nil {
			log.Printf("%s: Failed to save order - %s", r.RemoteAddr, err)
		}
	} else {
		log.Printf("%s: Payment failed", r.RemoteAddr)
		log.Printf("Response: %v", paymentResponse.Payment)
	}

	type orderResponse struct {
		PaymentStatus string `json:"payment_status"`
		NewOrderID    int    `json:"new_order_id"`
	}
	response := orderResponse{
		PaymentStatus: paymentResponse.Payment.Status,
		NewOrderID:    newID,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintf(w, string(responseJson))
}

func ReceiptPage(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	toView := make(map[string]interface{})

	fetchOrderFromURL(urlParams, toView)

	executeTemplate("receipt", w, toView)
}

func AllowedZips(w http.ResponseWriter, r *http.Request) {
	// This should be part of API..

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		jsonError(w, err)
		return
	}

	slotIds, found := params["slot-id"]
	if !found {
		jsonError(w, "SlotID not found")
		return
	}
	slotID := atoi(slotIds[0])

	var slot repositories.FulfillmentSlot
	err = repositories.FulfillmentSlotRepo.Get(persisters.ID(slotID), &slot)
	if err != nil {
		jsonError(w, "SlotID not found")
		return
	}

	allowedZips, err := json.Marshal(slot.ZipCodes)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, string(allowedZips))
}

func RegisterWebRoutes(cfg config.Config) http.Handler {
	serverConfiguration = cfg
	web := http.NewServeMux()

	web.HandleFunc("/order", OrderPage)
	web.HandleFunc("/delivery", DeliveryPage)
	web.HandleFunc("/review", ReviewPage)
	web.HandleFunc("/pay", PayPage)
	web.HandleFunc("/pickup", PickupPage)
	web.HandleFunc("/schedule", SchedulePage)
	web.HandleFunc("/submit", Submit)
	web.HandleFunc("/receipt", ReceiptPage)
	web.HandleFunc("/allowed-zips", AllowedZips)

	registerAPIRoutes(web)
	registerAdminRoutes(web)

	dir := http.Dir("./web-root/")
	fs := http.FileServer(dir)
	web.Handle("/", fs)

	return httpSubscribers(web)
}

func executeTemplate(name string, w http.ResponseWriter, data any) error {
	t := templates.ParseFiles("./pages/"+name+".html", "./pages/footer.html", "./pages/header.html")
	return t.ExecuteTemplate(w, name, data)
}

func fetchOrderFromURL(urlParams url.Values, toView map[string]interface{}) {
	var order repositories.Order
	err := json.Unmarshal([]byte(urlParams["order"][0]), &order)
	errorCheckHandleGraceful(err, toView)

	err = app.OrderManager.ReviewOrder(order)
	if err != nil {
		log.Println(err)
		toView["Errors"] = "There was an error with your order, please contact (775) 671-1945"
	} else {
		err = repositories.FillMenuItemDetails(&order)
		errorCheckHandleGraceful(err, toView)
		err = app.FillOrderFees(&order)
		errorCheckHandleGraceful(err, toView)
		toView["Order"] = order
		toView["OrderTotal"] = order.Total()
	}
}

func errorCheckHandleGraceful(err error, toView map[string]interface{}) {
	if err != nil {
		toView["Errors"] = err.Error()
		log.Println(err)
	}
}

func jsonError(w http.ResponseWriter, err interface{}) {
	w.WriteHeader(400)
	err = json.NewEncoder(w).Encode(err)
	if err != nil {
		log.Println(err)
	}
}
