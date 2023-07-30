package webserver

import "net/http"

func cachePolicy(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("cache-control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		handler.ServeHTTP(w, r)
	})
}
