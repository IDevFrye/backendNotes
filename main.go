package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func connectDB() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	fmt.Println("Connected to DB!")
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, text FROM notes ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Text)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func addNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO notes (text) VALUES ($1) RETURNING id", note.Text).Scan(&note.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func main() {
	connectDB()
	r := mux.NewRouter()

	r.HandleFunc("/notes", getNotes).Methods("GET")
	r.HandleFunc("/notes", addNote).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("Server is running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}