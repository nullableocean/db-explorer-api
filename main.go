package main

import (
	"bufio"
	"database/sql"
	"db_explorer/api"
	"db_explorer/dbexplorer"
	"db_explorer/pkg/router"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var ()

func main() {
	loadEnv()

	// DSN = "user:pass@tcp(localhost:3306)/dbname?charset=utf8"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)

	db, _ := sql.Open("mysql", dsn)
	err := db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	explorer := dbexplorer.NewSqlExplorer(db)
	controller := api.NewExplorerHandler(explorer)
	handler := router.NewMuxRouter()
	controller.RegisterRoutes(handler)

	port := os.Getenv("APP_PORT")

	fmt.Printf("server listen on http://localhost:%v\n", port)

	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalln(err)
	}
}

func loadEnv() {
	envFile, err := os.Open("./.env")
	if err != nil {
		log.Fatalln(err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		envVar := strings.Split(scanner.Text(), "=")
		os.Setenv(envVar[0], envVar[1])
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
