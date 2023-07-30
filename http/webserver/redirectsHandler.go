package webserver

import "net/http"

func redirectToSSL(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We may need an exception for certbot..
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		handler.ServeHTTP(w, r)
	})
}
