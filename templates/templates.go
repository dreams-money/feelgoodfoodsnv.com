package templates

import (
	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"
	"encoding/json"
	"log"
	"os"
)

func tp(path string) string {
	return "./pages/" + path
}

func generatePaymentScript(config config.Config) {
	file, err := os.Create("./web-root/assets/scripts/pay.js")
	logPanicOnError(err)

	scriptTemplate := ParseFiles(tp("assets/scripts/pay.js"))

	toScript := make(map[string]string)
	toScript["Environment"] = config.Environment
	toScript["ProdPublicPayKey"] = config.PublicPaymentKeys.Production
	toScript["DevPublicPayKey"] = config.PublicPaymentKeys.Development

	err = scriptTemplate.Execute(file, toScript)
	logPanicOnError(err)
}

func generateHTMLPage(pageSlug string, data interface{}) {
	file, err := os.Create("./web-root/" + pageSlug + ".html")
	logPanicOnError(err)

	htmlTemplate := ParseFiles(tp(pageSlug+".html"), tp("header.html"), tp("footer.html"))
	err = htmlTemplate.ExecuteTemplate(file, pageSlug, data)
	logPanicOnError(err)
}

func loadEnabledPages() []string {
	var pages []string
	file, err := os.ReadFile("./pages/enabled.json")
	logPanicOnError(err)

	json.Unmarshal(file, &pages)

	return pages
}

func everyTwoMinutes() {
	for _, page := range loadEnabledPages() {
		generateHTMLPage(page, nil)
	}
}

func atStart() {
	generatePaymentScript(serverConfiguration)
}

func logPanicOnError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
