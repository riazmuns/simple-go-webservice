package main

import (
	"fmt"
	"log"
	"net/http"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {

	// Configure the database connection (always check errors)
	// Specify connection properties.
	cfg := mysql.Config{
		User:   "root",
		Passwd: "",
		Addr:   "127.0.0.1:3306",
		DBName: "mysql",
	}

	// Get a driver-specific connector.
	connector, err := mysql.NewConnector(&cfg)
	fmt.Println(connector)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", "root@(127.0.0.1:3306)/mysql")

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//db := sql.OpenDB(connector)

	// Confirm a successful connection.
	// if err := db.Ping(); err != nil {
	// 	log.Fatal("?")
	// }

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

	// Handling static file
	fs := http.FileServer(http.Dir("static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("The server is listening at port 8090")
	http.ListenAndServe(":8090", r)
}
