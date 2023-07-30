package webserver

import (
	"net/http"
)

func test(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("sup"))
}

func registerAPIRoutes(server *http.ServeMux) {
	server.HandleFunc("/test", test)
}
