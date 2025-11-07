package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var templates *template.Template

func main() {
	templates = template.Must(template.ParseGlob("./src/template/*.html"))

	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("./src/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.HandleFunc("/", serveIndex)

	log.Println("Client Server starting on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"WebSocketURL": "ws://localhost:8080/ws",
	}
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}
}
