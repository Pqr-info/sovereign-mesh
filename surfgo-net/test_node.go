package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("SURFGO_LOCAL_TEST: Initializing S25 FE HUD...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body style='background:#000;color:#a9a9a9;font-family:monospace;'>")
		fmt.Fprintf(w, "<h1>SGO-LOCAL-NODE</h1><p>STATUS: STANDBY</p></body></html>")
	})
	fmt.Println("Local HUD active at http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)
}
