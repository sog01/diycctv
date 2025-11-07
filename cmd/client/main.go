package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", serveIndex)
	router.HandleFunc("/live", handleLiveStream)

	log.Println("Client Server starting on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}

func handleLiveStream(w http.ResponseWriter, r *http.Request) {
	live, _ := os.ReadFile("./src/live.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(live)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	index, _ := os.ReadFile("./src/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(index)
}
