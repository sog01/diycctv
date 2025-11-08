package main

import (
	"encoding/json"
	"html/template"
	"io"
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

	port := os.Getenv("CLIENT_PORT")
	log.Println("Client Server starting on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"WebSocketURL": os.Getenv("WEB_SOCKET_URL"),
	}
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}
}

func serveAdmin(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(os.Getenv("SERVER_BASE_URL") + "/api/live")
	if err != nil {
		log.Println("failed to hit server")
		return
	}
	defer resp.Body.Close()

	byt, _ := io.ReadAll(resp.Body)
	liveResp := make(map[string]any)
	if err := json.Unmarshal(byt, &liveResp); err != nil {
		log.Println("failed to unmarshal")
		return
	}
	data := []string{}
	for _, d := range liveResp["data"].([]any) {
		id, _ := d.(string)
		data = append(data, id)
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
