package main

import "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_facade"

func main() {
	apiFacade := api_facade.NewApiFacade()
	apiFacade.Start()
}
