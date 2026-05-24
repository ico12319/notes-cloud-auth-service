package middleware

import (
	"net/http"
	"sync"
)

func HTTPRequestTracker(wg *sync.WaitGroup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wg.Add(1)
			defer wg.Done()

			next.ServeHTTP(w, r)
		})
	}
}
