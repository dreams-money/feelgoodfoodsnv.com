package webserver

import (
	"fmt"
	"log"
	"net/http"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"
)

func ServeHTTP(cfg config.Config, sync chan struct{}) {
	switch cfg.Environment {
	case "dev":
		runDevServer(cfg, sync)
	case "prod":
		runProdServer(cfg, sync)
	default:
		log.Fatalln("Invalid environment: " + cfg.Environment)
	}
}

func runDevServer(cfg config.Config, sync chan struct{}) {
	httpPort := fmt.Sprintf(":%v", cfg.HttpPort)
	log.Println("Starting on Web on port " + httpPort)

	go func() {
		log.Println("Starting Dev Web Server")
		log.Fatal(http.ListenAndServe(httpPort, RegisterWebRoutes(cfg)))
		sync <- struct{}{}
	}()
}

func runProdServer(cfg config.Config, sync chan struct{}) {
	runDevServer(cfg, sync)
}

func httpSubscribers(server http.Handler) http.Handler {
	// Make dynamic based on config in future?  Or always code?
	// Backend can be opened at anytime.
	return logRequest(handleNotFound404(cachePolicy(handleGzipCompression(server))))
}
