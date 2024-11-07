package main

import (
	"fmt"
	"net/http"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! You accessed %s on %s\n", r.URL.Path, r.Host)
	})
	fmt.Println("Listening on :8081")
	http.ListenAndServe(":8081", handler)
}
