package main

import (
	"financial-app/internals/transaction/rest"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/transaction", rest.CreateTransaction).Methods("POST")
	r.HandleFunc("/transaction/{id}", rest.GetTransaction).Methods("GET")

	log.Println("Financial Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
