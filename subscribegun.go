package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	listName := r.URL.Path[1:]
	if len(listName) == 0 {
		http.Error(w, "No list specified!", 404)
		return
	}
	fmt.Fprintf(w, "Subscribe to: %s", listName)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
