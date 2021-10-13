package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/patrick246/partdb-csv/internal/auth"
	"github.com/patrick246/partdb-csv/internal/query"
	"github.com/patrick246/partdb-csv/internal/server"
	"log"
	"os"
	"strconv"
)

func main() {
	url := os.Getenv("MYSQL_URL")
	if url == "" {
		log.Fatalln("MYSQL_URL empty")
	}

	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Fatalf("connection error: %v", err)
	}

	querier := query.NewQuerier(db)
	partDbAuth := auth.NewPartDBAuthenticator(db)

	port, err := strconv.ParseUint(os.Getenv("PORT"), 10, 16)
	if err != nil {
		port = 8080
	}

	baseUrl := os.Getenv("PARTDB_BASEURL")
	if baseUrl == "" {
		log.Fatalln("missing env PARTDB_BASEURL")
	}

	srv := server.NewServer(
		uint(port),
		baseUrl,
		querier,
		partDbAuth,
	)
	log.Printf("listening port=%d", port)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
}
