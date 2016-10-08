package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
	"log"
	"net/http"
	"time"
)

var (
	port          = flag.String("port", "80", "The application port")
	dbinfo        = flag.String("dbinfo", "host=recording_db user=postgres password=postgres dbname=postgres sslmode=disable", "Database connection string")
	migrationsDir = flag.String("migrations", "/migrations/", "Path to the migrations directory")
)

var db *sql.DB

type recording struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

func index(w http.ResponseWriter, _ *http.Request) {
	query := "SELECT id, url FROM recordings"
	rows, err := db.Query(query)
	check(err)

	res := make([]recording, 0)
	for rows.Next() {
		var rec recording
		err := rows.Scan(&rec.ID, &rec.URL)
		check(err)
		res = append(res, rec)
	}
	err = rows.Err()
	check(err)

	json.NewEncoder(w).Encode(res)
}

func create(w http.ResponseWriter, r *http.Request) {
	var rec recording
	err := json.NewDecoder(r.Body).Decode(&rec)
	check(err)

	_, err = db.Exec("INSERT INTO recordings (id, url) VALUES($1, $2)", rec.ID, rec.URL)
	check(err)

	w.WriteHeader(http.StatusCreated)
}

func serve(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		index(w, r)
	} else if r.Method == http.MethodPut {
		create(w, r)
	}
}

func main() {
	flag.Parse()

	var err error
	db, err = sql.Open("postgres", *dbinfo)
	check(err)
	defer db.Close()

	for {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: *migrationsDir,
	}
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Applied %d migrations!\n", n)
	}

	http.HandleFunc("/", serve)
	log.Printf("Starting rest service on port :%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
