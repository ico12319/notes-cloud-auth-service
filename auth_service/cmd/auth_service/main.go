package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type output struct {
	Text string `json:"text"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/authService/api/v1/hello-world", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(&output{
			Text: "Hello world",
		}); err != nil {
			http.Error(w, "failed to encode JSON response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server started on http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
