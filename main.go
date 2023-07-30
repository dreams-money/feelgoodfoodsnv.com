package main

import (
	"flag"
	"log"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/app"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/http/webserver"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/templates"
)

func main() {
	jsonConfigFile := parseCommandLineOptions()
	if jsonConfigFile == "" {
		jsonConfigFile = "config.json"
	}
	config := config.LoadConfig(jsonConfigFile)

	syncChannel := make(chan struct{})

	webserver.ServeHTTP(config, syncChannel)
	app.ProcessOrders(config, syncChannel)
	templates.RunTemplateOutputInterval(config, syncChannel)

	for {
		<-syncChannel
	}
}

func parseCommandLineOptions() string {
	flag.Parse()
	arguments := flag.Args()
	if len(arguments) > 1 {
		log.Fatalf("Usage: feelgood-server <config file path>")
	}
	if len(arguments) == 0 {
		return ""
	}

	return arguments[0]
}
