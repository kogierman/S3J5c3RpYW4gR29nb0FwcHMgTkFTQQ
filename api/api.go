package api

import (
	"fmt"
	"log"
	"net/http"

	"example.com/nasa-apod-fetcher/handlers"
	"example.com/nasa-apod-fetcher/nasa"
	"github.com/gorilla/mux"
)

type API struct {
	apod *nasa.APOD
}

func NewAPI(apod *nasa.APOD) *API {
	return &API{apod}
}

func (a *API) Listen(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/pictures", handlers.NewPictures(a.apod).Handle).Methods("GET")

	log.Printf("Start listening on port: %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
