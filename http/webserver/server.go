package webserver

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"

	"golang.org/x/crypto/acme/autocert"
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
	httpPort := ":80"
	log.Println("Starting on Web on port " + httpPort)

	go func() {
		log.Println("Starting Dev Web Server")
		log.Fatal(http.ListenAndServe(httpPort, RegisterWebRoutes(cfg)))
		sync <- struct{}{}
	}()
}

func runProdServer(cfg config.Config, sync chan struct{}) {
	// Automated SSL certs using Let's Encrypt HTTP-01
	// Multi porting an application will require DNS-01 challenge..
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Hosts...),
		Cache:      autocert.DirCache("certs"),
	}

	httpsPort := fmt.Sprintf(":" + strconv.Itoa(cfg.TLSPort))
	webServer := &http.Server{
		Addr:      httpsPort,
		TLSConfig: certManager.TLSConfig(),
		Handler:   RegisterWebRoutes(cfg),
	}

	go func() {
		log.Fatal(webServer.ListenAndServeTLS("", "")) // HTTPS
		sync <- struct{}{}
	}()

	webPort := fmt.Sprintf(":" + strconv.Itoa(cfg.HttpPort))
	handler := certManager.HTTPHandler(RegisterWebRoutes(cfg))
	go func() {
		log.Fatal(http.ListenAndServe(webPort, redirectToSSL(handler))) // Web
		sync <- struct{}{}
	}()

	m := "Serving: %+v on ports ssl:%v http:%v"
	log.Printf(m, cfg.Hosts, httpsPort, webPort)
}

func httpSubscribers(server http.Handler) http.Handler {
	// Make dynamic based on config in future?  Or always code?
	// Backend can be opened at anytime.
	return logRequest(handleNotFound404(cachePolicy(handleGzipCompression(server))))
}
