package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// TYPES
type Client struct {
	ID_client int    `json:"id_client"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

var (
	db        *sql.DB
	clientsMu sync.RWMutex
	nextID    = 1
)

func main() {
	/// BASE DE DONNÉES
	var err error
	db, err = sql.Open("mysql", "goteam:root@tcp(localhost:3306)/golang")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	port := 8080
	fmt.Printf("Server is running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

	/// TABLES
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS clients (
			id_client INT AUTO_INCREMENT PRIMARY KEY,
			firstname VARCHAR(150),
			lastname VARCHAR(150),
			email VARCHAR(150),
			password VARCHAR(255)
		);
    `)
	if err != nil {
		log.Fatal(err)
	}

	/// ROUTES

	/// Clients
	http.HandleFunc("/api/clients", getClientsHandler)
	http.HandleFunc("/api/clients/add", addClientHandler)
	http.HandleFunc("/api/clients/update", updateClientHandler)
	http.HandleFunc("/api/clients/delete", deleteClientHandler)
}

//FUNCTIONS

// CLIENTS
func addClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var newClient Client
	err := json.NewDecoder(r.Body).Decode(&newClient)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO clients (firstname, lastname, email, password) VALUES (?, ?, ?, ?)", newClient.Firstname, newClient.Lastname, newClient.Email, newClient.Password)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newClient.ID_client = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newClient)
}

func getClientsHandler(w http.ResponseWriter, r *http.Request) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	// Fetch users from the database
	rows, err := db.Query("SELECT * FROM clients")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var clientList []Client
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.ID_client, &client.Firstname, &client.Lastname, &client.Email, &client.Password)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		clientList = append(clientList, client)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientList)
}

func updateClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var updatedClient Client
	err := json.NewDecoder(r.Body).Decode(&updatedClient)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientsMu.RLock()
	defer clientsMu.RUnlock()
	row := db.QueryRow("SELECT id_client FROM clients WHERE id_client=?", updatedClient.ID_client)
	if err := row.Scan(&updatedClient.ID_client); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE clients SET firstname=?, lastname=?, email=?, password=? WHERE id_client=?", updatedClient.Firstname, updatedClient.Lastname, updatedClient.Email, updatedClient.Password, updatedClient.ID_client)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedClient)
}

func deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Query().Get("id_client")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()
	row := db.QueryRow("SELECT id_client FROM clients WHERE id_client=?", id)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM clients WHERE id_client=?", id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
