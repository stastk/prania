package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"main/models"

	_ "github.com/lib/pq"
)

func main() {
	var err error

	// Initalize the sql.DB connection pool and assign it to the models.DB
	// global variable.
	models.DB, err = sql.Open("postgres", "user=gouser password=gopass dbname=prania_exp sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/initdb", initDB)
	http.HandleFunc("/dropdb", dropDB)
	http.HandleFunc("/ingridients", ingridientsIndex)
	http.HandleFunc("/ingridients_variations", ingvarsIndex)
	http.ListenAndServe(":3000", nil)
}

// ingridientsIndex sends a HTTP response listing all ingridients.
func dropDB(w http.ResponseWriter, r *http.Request) {
	answer, err := models.DropDB()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, answer := range answer {
		fmt.Fprintf(w, "%s\n", answer)
	}
}

// ingridientsIndex sends a HTTP response listing all ingridients.
func initDB(w http.ResponseWriter, r *http.Request) {
	answer, err := models.InitDB()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, answer := range answer {
		fmt.Fprintf(w, "%s\n", answer)
	}
}

// ingridientsIndex sends a HTTP response listing all ingridients.
func ingridientsIndex(w http.ResponseWriter, r *http.Request) {
	ings, err := models.AllIngridients()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, ing := range ings {
		fmt.Fprintf(w, "%d : %s\n", ing.Id, ing.Name)
	}
}

// ingridientsIndex sends a HTTP response listing all ingridients variations.
func ingvarsIndex(w http.ResponseWriter, r *http.Request) {
	ingvars, err := models.AllIngvars()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, ingvar := range ingvars {
		fmt.Fprintf(w, "%d : %s\n", ingvar.Id, ingvar.Name)
	}
}