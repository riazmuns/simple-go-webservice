package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

func insertItem(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Item struct
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser) // the post body must comply with User struct

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the item into the database
	res, err := db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		newUser.Username,
		newUser.Password,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the inserted item
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the ID of the inserted item in the response
	response := map[string]int{"id": int(lastInsertID)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func main() {

	// Configure the database connection (always check errors)
	// Specify connection properties.
	cfg := mysql.Config{
		User:      "root",
		Passwd:    "",
		Addr:      "127.0.0.1:3306",
		DBName:    "mysql",
		Collation: "utf8mb4_general_ci",
	}

	// Get a driver-specific connector.
	conn, err := mysql.NewConnector(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	// Open a connection to the database using the connector
	db = sql.OpenDB(conn)
	defer db.Close()

	//db, err := sql.Open("mysql", "root@(127.0.0.1:3306)/mysql") -> this is another way to make the call

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// we need to make sure that this code creates a table if it doesn't exist
	var tableExists bool

	// checking if table exists
	query := `SHOW TABLES LIKE 'users';`
	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	} else {
		tableExists = true
	}

	// Note: there is a bug here - i.e once the table is created and dropped, it still says table exists
	fmt.Println(tableExists)

	if !tableExists { // Create a new table if it doesn't exist
		query := `
			CREATE TABLE users(
				id INT AUTO_INCREMENT,
				username TEXT NOT NULL,
				password TEXT NOT NULL,
				created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
				PRIMARY KEY (id)

			);`

		if _, err := db.Exec(query); err != nil {
			log.Fatal(err)
		}
	}

	r := mux.NewRouter()

	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		// get the book
		// navigate to the page
		vars := mux.Vars(r)
		fmt.Fprintf(w, "Title %q Page %q", vars["title"], vars["page"])
	})

	r.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello you requested %q", req.URL)
	})

	// Define your routes
	r.HandleFunc("/items", insertItem).Methods("POST")

	// Handling static file
	fs := http.FileServer(http.Dir("static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("The server is listening at port 8090")
	http.ListenAndServe(":8090", r)
}
