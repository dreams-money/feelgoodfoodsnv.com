package webserver

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/app"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/templates"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

var adminToken string

func login(resp http.ResponseWriter) {
	http.SetCookie(resp, &http.Cookie{
		Name:   "api-token",
		Value:  adminToken,
		MaxAge: 86400,
	})
}

func isLoggedIn(req *http.Request) bool {
	cookie, err := req.Cookie("api-token")
	if err != nil || cookie.Value != adminToken {
		log.Println(req.RemoteAddr + " Unauthorized attempt to access admin area")
		return false
	}

	return true
}

// We need a timed cached to track login attempts..
// Self expiring, because else we exposed memory leak possibility
func adminLogin(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		if req.PostFormValue("email") == "admin@feelgoodfoodsnv.com" &&
			req.PostFormValue("password") == "asdfasdf" {
			login(resp)
			log.Println(req.RemoteAddr + " Admin login")
			http.Redirect(resp, req, "/admin/home/", 302)
			return
		} else {
			log.Println(req.RemoteAddr + " Admin failed login attempt")
		}
	}

	t := templates.ParseFiles("./pages/admin/index.html", "./pages/footer.html", "./pages/header.html")
	t.ExecuteTemplate(resp, "admin/index", nil)
}

func adminHomeIndex(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	t := templates.ParseFiles("./pages/admin/home.html", "./pages/footer.html", "./pages/admin/header.html")
	t.ExecuteTemplate(resp, "admin/home", nil)
}

func adminMenuItem(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	var menuItem repositories.MenuItem
	menuItem.NutritionLabel = &repositories.NutritionLabel{}

	toView := make(map[string]interface{})
	repo := repositories.MenuItemRepo

	var err error
	id := 0 // 0 means new
	reqEditId := req.URL.Query().Get("edit")
	if reqEditId != "" {
		id, err = strconv.Atoi(reqEditId)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		err = repo.Get(persisters.ID(id), &menuItem)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}
	}

	if req.Method == http.MethodPost {
		err := req.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		nutritionLabel := &repositories.NutritionLabel{
			Calories:      atoi(req.Form["cals"][0]),
			Carbohydrates: atoi(req.Form["carbs"][0]),
			Protiens:      atoi(req.Form["protiens"][0]),
			Fats:          atoi(req.Form["fats"][0]),
		}

		menuItem = repositories.MenuItem{
			Name:           req.Form["name"][0],
			Description:    req.Form["description"][0],
			Price:          atof(req.Form["price"][0]),
			Category:       req.Form["category"][0],
			Qualities:      req.Form["qualities"],
			Image:          menuItem.Image,
			NutritionLabel: nutritionLabel,
			Timestamps:     repositories.Now(),
		}

		if menuItem.Name == "" {
			log.Println("Empty form; needs javascript")
			http.Error(resp, "Menu item needs name at least", 500)
			return
		}

		image, _, err := req.FormFile("image")
		if image != nil && err == nil {
			defer image.Close()

			imageBytes, err := ioutil.ReadAll(image)
			if err != nil {
				log.Println(err.Error())
				http.Error(resp, "", 500)
				return
			}

			menuItem.Image = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageBytes)
		}

		id, err := repo.Set(persisters.ID(id), &menuItem)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		menuItem.ID = id

		if reqEditId == "" { // Redirect to correct URL if brand new object
			http.Redirect(resp, req, "/admin/menu-item/?edit="+strconv.Itoa(int(id)), 303)
			return
		}
	}

	toView[menuItem.Category] = true
	for _, quality := range menuItem.Qualities {
		toView[quality] = true
	}

	toView["menuItem"] = menuItem
	toView["listing"] = repo.List()

	t := templates.ParseFiles("./pages/admin/menuItem.html", "./pages/footer.html", "./pages/admin/header.html")
	err = t.ExecuteTemplate(resp, "admin/menuItem", toView)

	if err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		_ = t.ExecuteTemplate(resp, "error", nil)
		return
	}
}

func adminMenu(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	if req.Method == http.MethodPost {
		req.ParseForm()

		var activeIDs []persisters.ID
		activeItemSubmitted := req.Form["active_menu"]
		for _, item := range activeItemSubmitted {
			id, err := strconv.Atoi(item)
			if err != nil {
				log.Println(err)
			}
			activeIDs = append(activeIDs, persisters.ID(id))
		}

		repositories.MenuItemRepo.SetActive(activeIDs)
	}

	toView := make(map[string]interface{})

	toView["menuItems"] = repositories.MenuItemRepo.ListWithDetails()

	t := templates.ParseFiles("./pages/admin/menu.html", "./pages/footer.html", "./pages/admin/header.html")
	err := t.ExecuteTemplate(resp, "admin/menu", toView)

	if err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		_ = t.ExecuteTemplate(resp, "error", nil)
		return
	}
}

func adminCustomer(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	var billingAddress repositories.Address
	var customer repositories.Customer
	toView := make(map[string]interface{})

	repo := repositories.CustomerRepo

	var err error
	id := 0 // 0 means new
	reqEditId := req.URL.Query().Get("edit")
	if reqEditId != "" {
		id, err = strconv.Atoi(reqEditId)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		err = repo.Get(persisters.ID(id), &customer)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}
	}

	if req.Method == http.MethodPost {
		err := req.ParseForm()
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		billingAddress = repositories.Address{
			Line1:      req.Form["address1"][0],
			Line2:      req.Form["address2"][0],
			City:       req.Form["city"][0],
			State:      req.Form["state"][0],
			Postal:     req.Form["postal"][0],
			Timestamps: repositories.Now(),
		}
		customer = repositories.Customer{
			FirstName:  req.Form["first"][0],
			LastName:   req.Form["last"][0],
			Phone:      req.Form["phone"][0],
			Email:      req.Form["email"][0],
			Addresses:  []repositories.Address{billingAddress},
			Timestamps: repositories.Now(),
		}

		if customer.FirstName == "" {
			log.Println("Empty form; needs javascript")
			http.Error(resp, "Customer needs first name at least", 500)
			return
		}

		id, err := repo.Set(persisters.ID(id), customer)
		if err != nil {
			log.Println(err.Error())
			http.Error(resp, "", 500)
			return
		}

		customer.ID = id

		if reqEditId == "" { // Redirect to correct URL if brand new object
			http.Redirect(resp, req, "/admin/customer/?edit="+strconv.Itoa(int(id)), 303)
			return
		}
	}

	toView["customer"] = customer
	if len(customer.Addresses) > 0 {
		toView["street"] = customer.Addresses[0].Line1
		toView["city"] = customer.Addresses[0].City
		toView["state"] = customer.Addresses[0].State
		toView["postal"] = customer.Addresses[0].Postal
	}
	toView["listing"] = repo.List()

	t := templates.ParseFiles("./pages/admin/customer.html", "./pages/footer.html", "./pages/admin/header.html")
	err = t.ExecuteTemplate(resp, "admin/customer", toView)

	if err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		_ = t.ExecuteTemplate(resp, "error", nil)
		return
	}
}

func adminSlot(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	var slot repositories.FulfillmentSlot
	repo := repositories.FulfillmentSlotRepo

	id := 0 // 0 means new
	var err error
	reqEditId := req.URL.Query().Get("edit")
	if reqEditId != "" {
		id, err = strconv.Atoi(reqEditId)
		if inErrorThenLogAndReportAndHalt(err != nil, errMsg(err), "", resp) {
			return
		}

		err = repo.Get(persisters.ID(id), &slot)
		if inErrorThenLogAndReportAndHalt(err != nil, errMsg(err), "", resp) {
			return
		}
	}

	if req.Method == http.MethodPost {
		err := req.ParseForm()
		if inErrorThenLogAndReportAndHalt(err != nil, errMsg(err), "", resp) {
			return
		}

		var fee *repositories.Fee
		feeAmount := atof(req.Form["fee_amount"][0])
		feeName := req.Form["fee_name"][0]
		if feeAmount > 0 && feeName != "" {
			fee = &repositories.Fee{
				Amount: feeAmount,
				Name:   feeName,
			}
		} else {
			fee = nil
		}

		var zipCodes []int
		zipCodesInput := req.Form["zip_codes"][0]
		if zipCodesInput != "" {
			zipCodesInput = strings.ReplaceAll(zipCodesInput, " ", "")
			inputZipCodes := strings.Split(zipCodesInput, ",")
			for _, zipCode := range inputZipCodes {
				zipCodes = append(zipCodes, atoi(zipCode))
			}
		}

		slot = repositories.FulfillmentSlot{
			DayOfWeek:       atoi(req.Form["day"][0]),
			SlotDescription: req.Form["slot_description"][0],
			MaxFils:         atoi(req.Form["max_fills"][0]),
			Fee:             fee,
			ZipCodes:        zipCodes,
		}
		if inErrorThenLogAndReportAndHalt(slot.SlotDescription == "",
			"Admin slot needs description at least", "Empty form; needs javascript", resp) {
			return
		}

		id, err := repo.Set(persisters.ID(id), &slot)
		if inErrorThenLogAndReportAndHalt(err != nil, errMsg(err), "", resp) {
			return
		}

		slot.ID = id

		if reqEditId == "" { // Redirect to correct URL if brand new object
			http.Redirect(resp, req, "/admin/slot/?edit="+strconv.Itoa(int(id)), 303)
			return
		}
	}

	toView := make(map[string]interface{})
	toView["slot"] = slot
	toView["dayofweek"] = slot.DayOfWeek
	toView["listing"] = repo.List()

	t := templates.ParseFiles("./pages/admin/slot.html", "./pages/footer.html", "./pages/admin/header.html")
	err = t.ExecuteTemplate(resp, "admin/slot", toView)
	ifErrorSendToErrorPage(resp, t, err)
}

func adminOrderSheet(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	toView := make(map[string]interface{})
	toView["OrderSheet"] = app.CreateOrderSheet()
	t := templates.ParseFiles("./pages/admin/orders.html", "./pages/footer.html", "./pages/admin/header.html")
	err := t.ExecuteTemplate(resp, "admin/orders", toView)
	ifErrorSendToErrorPage(resp, t, err)
}

func adminDeliverySheet(resp http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(resp, req, "/admin/", 307)
		return
	}

	toView := make(map[string]interface{})
	toView["DeliverySheet"] = app.CreateDeliverySheet()
	t := templates.ParseFiles("./pages/admin/deliveries.html", "./pages/footer.html", "./pages/admin/header.html")
	err := t.ExecuteTemplate(resp, "admin/deliveries", toView)
	ifErrorSendToErrorPage(resp, t, err)
}

func initLoginToken() {
	loginLokenPath := "data/login-token"
	_, err := os.Stat(loginLokenPath)
	if errors.Is(err, os.ErrNotExist) {
		tokenGenerator := func(n int) string {
			var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

			b := make([]rune, n)
			for i := range b {
				b[i] = letters[rand.Intn(len(letters))]
			}
			return string(b)
		}

		_, err := os.Stat("data")
		if errors.Is(err, os.ErrNotExist) {
			os.Mkdir("data", 0700)
		}

		adminToken = tokenGenerator(75)
		err = os.WriteFile(loginLokenPath, []byte(adminToken), 0700)
		if err != nil {
			log.Panic(err.Error())
		}
	} else {
		var token []byte
		token, err = os.ReadFile(loginLokenPath)
		if err != nil {
			log.Panic(err.Error())
		}
		adminToken = string(token)
	}
}

func registerAdminRoutes(server *http.ServeMux) {
	initLoginToken()

	server.HandleFunc("/admin/", adminLogin)
	server.HandleFunc("/admin/home/", adminHomeIndex)
	server.HandleFunc("/admin/menu-item/", adminMenuItem)
	server.HandleFunc("/admin/menu/", adminMenu)
	server.HandleFunc("/admin/customer/", adminCustomer)
	server.HandleFunc("/admin/slot/", adminSlot)
	server.HandleFunc("/admin/orders/", adminOrderSheet)
	server.HandleFunc("/admin/deliveries/", adminDeliverySheet)
}

func inErrorThenLogAndReportAndHalt(condition bool,
	logMessage, userMessage string, resp http.ResponseWriter) bool {
	if condition {
		log.Println(logMessage)
		http.Error(resp, userMessage, 500)
	}

	return condition
}

func ifErrorSendToErrorPage(resp http.ResponseWriter, t *template.Template, err error) {
	if err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		_ = t.ExecuteTemplate(resp, "error", nil)
		return
	}
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
	}
	return i
}

func atof(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		log.Println(err)
	}
	return float32(f)
}

func errMsg(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
