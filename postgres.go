package main

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func writeToPostgres(host, username, password, port, database, sslMode string, trackData []SpotifyTrack) {
	// Connect to PG
	db, err := connectToDatabase(host, username, password, port, database, sslMode)
	if err != nil {
		log.Fatalf("Failed to connect to pg: %v", err)
	}
	tableName := "trackHistory2024"

	// Create or use table
	tableExists, err := tableExists(db, tableName)
	if err != nil {
		log.Fatalf("Failed to assert pg table exists: %v", err)
	}
	if !tableExists {
		err := createTable(db, tableName)
		if err != nil {
			log.Fatalf("Failed to create pg table: %v", err)
		}
	}

	// Sort tracks oldest to newest
	sortedTracks := sortTracks(trackData)

	// Prep data for insert to PG
	for _, track := range sortedTracks {
		err = insertTrack(db, tableName, track)
		if err != nil {
			pqErr, ok := err.(*pq.Error)
			if !ok || pqErr.Code != "23505" { // If not a postgres error, or if a postgres error and not duplicate key constraint violation
				postgresErrors.Inc()
				log.Errorf("Failed to insert track %v %v: %v", track.URL, track.PlayedAt, err)
			}
		}
	}
}

func insertTrack(db *sql.DB, tableName string, track SpotifyTrack) error {
	query := `
        INSERT INTO ` + tableName + ` (playedat, name, artist, url) VALUES ($1, $2, $3, $4)
    `
	_, err := db.Exec(query, track.PlayedAt, track.Name, track.Author, track.URL)
	if err != nil {
		return err
	}

	return nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM   information_schema.tables
            WHERE  table_schema = 'public'
            AND    table_name = $1
        )
    `

	var exists bool
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func connectToDatabase(host, user, password, port, dbname, sslmode string) (*sql.DB, error) {
	// Build the connection string
	connectionString :=
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname)

	// Open a connection to the database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Info("Connected to the database")
	return db, nil
}

func createTable(db *sql.DB, tableName string) error {
	query := `
        CREATE TABLE IF NOT EXISTS ` + tableName + ` (
            name VARCHAR(255) NOT NULL,
            artist VARCHAR(255) NOT NULL,
            url VARCHAR(255) NOT NULL,
			playedat TIMESTAMP NOT NULL,
			PRIMARY KEY(url, playedat)
        )
    `

	_, err := db.Exec(query)
	return err
}
