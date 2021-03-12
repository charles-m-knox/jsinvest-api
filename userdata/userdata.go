package userdata

import (
	"database/sql"
	"encoding/json"
	"fa-middleware/config"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func SetUserData(conf config.Config, id string, userData interface{}) error {
	// https://pkg.go.dev/github.com/lib/pq?utm_source=godoc#hdr-Bulk_imports
	connStr := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?%v",
		conf.PostgresUser,
		conf.PostgresPass,
		conf.PostgresHost,
		conf.PostgresPort,
		conf.PostgresDBName,
		conf.PostgresOptions,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %v", err.Error())
	}

	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err.Error())
	}

	// Work in progress: doesn't quite work
	rows, err := db.Query(
		`CREATE TABLE IF NOT EXISTS userdata  (
			id VARCHAR ( 36 ) PRIMARY KEY,
			user_data_json TEXT
		);`,
	)
	if err != nil {
		return fmt.Errorf("failed to perform create table query: %v", err.Error())
	}

	log.Printf("create table rows: %v", rows)

	// Work in progress: doesn't quite work
	rows, err = db.Query(
		fmt.Sprintf(
			`INSERT INTO userdata(id, user_data_json) VALUES(%v, %v) RETURNING id;`,
			id,
			string(userDataBytes),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err.Error())
	}

	log.Printf("insert rows: %v", rows)

	return nil

}
