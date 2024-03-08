package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/riazmuns/simple-go-webservice/survey"
	"github.com/riazmuns/simple-go-webservice/todo"

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

// creating a TODO template
func todoTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.html"))

	data := todo.TodoPageData{
		PageTitle: "My TODO list",
		Todos: []todo.Todo{
			{Title: "Task 1", Done: false},
			{Title: "Task 2", Done: true},
			{Title: "Task 3", Done: true},
		},
	}

	tmpl.Execute(w, data)
}

// handler for survey
func showSurvey(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/form.html"))
	// checking if the http request is post method
	// TODO: maybe its better to tell the user, this endpoint is not supported or bad request
	if r.Method == http.MethodPost {
		// Inticate that the form was submitted

		// Populating the values from the form to render the filled values
		details := survey.ContactDetails{
			Email:   r.FormValue("email"),
			Subject: r.FormValue("subject"),
			Message: r.FormValue("message"),
		}

		fmt.Println(details)

		// TODO: persist the survey data into a database

		tmpl.Execute(w, struct{ Success bool }{true})
		return
	}

	tmpl.Execute(w, struct{ Success bool }{false})
}

func insertUser(w http.ResponseWriter, r *http.Request) {
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

/*
Search by ID
*/
func queryUser(w http.ResponseWriter, r *http.Request) {

	// Get the id parameter from the URL
	params := mux.Vars(r)
	userID := params["id"]

	fmt.Println(params)

	var queryUser User

	// making a select query - scan will populate the values from the select query
	err := db.QueryRow("SELECT id, username FROM users WHERE id=?", userID).Scan(&queryUser.ID, &queryUser.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		// handle rest of the errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d, "username": %s}`, queryUser.ID, queryUser.Username)

}

/*
Query user from URL param
If ID is passed we will select a single record
If Multiple record matches, we will return all the matches
*/

func queryUserFromURL(w http.ResponseWriter, r *http.Request) {

	// Get the id parameter from the URL
	query := r.URL.Query()

	userID := query.Get("id")
	username := query.Get("username")

	var queryUser User

	// checking is userID is define to select only one record
	if userID != "" {
		// making a select query - scan will populate the values from the select query
		err := db.QueryRow("SELECT id, username FROM users WHERE id=?", userID).Scan(&queryUser.ID, &queryUser.Username)

		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			// handle rest of the errors
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": %d, "username": %s}`, queryUser.ID, queryUser.Username)

	} else if username != "" {

		rows, err := db.Query("SELECT id, username, password FROM users WHERE username=?", username)

		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			// handle rest of the errors
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		defer rows.Close()

		var users []User
		// Populating the matched records from the db
		for rows.Next() {
			var u User

			err := rows.Scan(&u.ID, &u.Username, &u.Password)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			users = append(users, u)
		}

		w.WriteHeader(http.StatusOK)

		// we need to return as a json
		jsonData, err := json.Marshal(users)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Setting the response header
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	} else {
		// this is not a valid request
		http.Error(w, "Bad Request", http.StatusInternalServerError)
	}

}

func removeUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	rows, err := db.Exec("DELETE FROM users WHERE  id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	cnt, err := rows.RowsAffected()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if cnt == 1 {
		// making delete more verbose so that we have return the ID of what was deleted
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": %s, "deleted": True}`, userID)

	} else {
		http.Error(w, "Critical: ID is not unique", http.StatusInternalServerError)
	}
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
	//fmt.Println(tableExists)

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

	// Assert a New User - Create
	r.HandleFunc("/newUser", insertUser).Methods("POST")

	// Find a User - Read - note it needs the regex
	// Example: localhost:8090/user/1
	r.HandleFunc("/user/{id:[0-9]+}", queryUser).Methods("GET")

	// Find a User using params
	r.HandleFunc("/user", queryUserFromURL).Methods("GET")

	// Delete a User using ID
	r.HandleFunc("/removeUser/{id:[0-9]+}", removeUser).Methods("DELETE")

	// Survery Form - we need Get to render the UI, and Post to handle submit
	r.HandleFunc("/survey", showSurvey).Methods("POST", "GET")

	// Handling static file
	fs := http.FileServer(http.Dir("static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	// Todo template
	r.HandleFunc("/todo", todoTemplate).Methods("GET")

	fmt.Println("The server is listening at port 8090")
	http.ListenAndServe(":8090", r)
}
