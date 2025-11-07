package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
)

var templates *template.Template

func main() {
	gotenv.Load()
	templates = template.Must(template.ParseGlob("./src/template/*.html"))

	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("./src/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.HandleFunc("/", serveIndex)
	router.HandleFunc("/admin", serveAdmin)
	router.HandleFunc("/stream/{id}", serveStream)

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

func serveAdmin(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("./records")
	if err != nil {
		w.Write([]byte("error"))
		return
	}

	data := []string{}
	for _, file := range files {
		data = append(data, file.Name())
	}

	err = templates.ExecuteTemplate(w, "admin.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}
}

func serveStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	data := map[string]any{
		"WebSocketURL": os.Getenv("WEB_SOCKET_URL"),
		"Id":           vars["id"],
	}
	err := templates.ExecuteTemplate(w, "stream.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}
}
