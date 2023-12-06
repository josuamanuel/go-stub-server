package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Card struct {
	Path  string `json:"path"`
	Delay int    `json:"delay"`
	JSON  struct {
		Person struct {
			Name string `json:"name"`
			DOB  string `json:"DOB"`
		} `json:"person"`
		Active bool `json:"active"`
	} `json:"JSON"`
}

func main() {
	// Read the JSON file into memory
	data, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the JSON data
	var cards []Card
	err = json.Unmarshal(data, &cards)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP server
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Define the handler function for the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Find the card that matches the requested path
		var card *Card
		for _, c := range cards {
			if strings.Contains(r.URL.Path, c.Path) {
				card = &c
				break
			}
		}

		// If no matching card is found, return a 404 error
		if card == nil {
			http.NotFound(w, r)
			return
		}

		// Delay the response if specified
		time.Sleep(time.Duration(card.Delay) * time.Millisecond)

		// Set the content type header to JSON
		w.Header().Set("Content-Type", "application/json")

		// Marshal the card data to JSON
		response, err := json.Marshal(card.JSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		fmt.Fprint(w, string(response))
	})

	// Start the server
	log.Println("Server listening on port 8080")
	log.Fatal(server.ListenAndServe())
}
