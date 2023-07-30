package webserver

import (
	"log"
	"net/http"
)

type notFoundHandler struct {
	http.ResponseWriter
	status int
}

func (w *notFoundHandler) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundHandler) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

func handleNotFound404(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFound := &notFoundHandler{ResponseWriter: w}
		handler.ServeHTTP(notFound, r)
		if notFound.status == 404 {
			log.Printf("%s %s %s HTTP/404\n", r.RemoteAddr, r.Method, r.URL)

			// Implement network policy here
			// I.e. prevent bots from pinging the site for a time.

			http.Redirect(w, r, "/404.html", http.StatusFound)
		}
	})
}
