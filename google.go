package main

import (
	"context"
	"encoding/base64"
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func sortTracks(tracks []SpotifyTrack) []SpotifyTrack {
	// Sort tracks by PlayedAt in ascending order (oldest first)
	sort.Slice(tracks, func(i, j int) bool {
		return tracks[i].PlayedAt.Before(tracks[j].PlayedAt)
	})

	return tracks
}

func writeToSheet(spreadsheetID string, sheetName string, tracksToInsert []SpotifyTrack) {
	// Authenticate with Google Sheets API
	sheetsService, err := getSheetsService()
	if err != nil {
		log.Fatal("Error authenticating with Google Sheets API: ", err)
	}

	var values [][]interface{}

	// Sort tracks oldest to newest
	sortedTracks := sortTracks(tracksToInsert)

	// Iterate through tracks and add them to the values
	i := 0
	for _, track := range sortedTracks {
		i++
		log.Infof("Iteration: %v", i)

		// Search for any matches in the Google Sheet
		if matchExists(track) {
			log.Infof("Track already exists in Google Sheet. Skipping insertion: %+v", track)
			continue
		}

		// If the row doesn't exist or doesn't match the SpotifyTrack, add it to the values
		values = append(values, []interface{}{track.Name, track.Author, track.URL, track.PlayedAt.Format(timeLayout)})

		// Add more logging to identify where the issue might be
		log.Infof("Added new track to values: %+v", track)
	}

	// If no new entries, exit early
	if len(values) == 0 {
		log.Info("No new entries to write.")
		return
	}

	writeRange := sheetName + "!A1" // Adjust the range as needed

	_, err = sheetsService.Spreadsheets.Values.Append(spreadsheetID, writeRange, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("RAW").Do()

	if err != nil {
		log.WithFields(log.Fields{
			"spreadsheetID": spreadsheetID,
			"sheetName":     sheetName,
			"error":         err,
		}).Fatal("Error writing to Google Sheet")
	} else {
		log.Info("Data written to Google Sheet successfully.")
	}
}

func getSheetsService() (*sheets.Service, error) {
	ctx := context.Background()

	encodedCreds := os.Getenv("GOOGLE_SHEETS_CREDENTIAL")
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		log.Fatal("Error decoding base64-encoded credentials: ", err)
	}

	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsJSON(decodedCreds), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}

	return sheetsService, nil
}

var globalSheetData *sheets.ValueRange

// Modify getExistingTracksFromSheet function to use local data
func getExistingTracksFromSheet(spreadsheetID, sheetName string) ([]SpotifyTrack, error) {
	if globalSheetData == nil {
		// Fetch sheet data only if not already fetched
		sheetsService, err := getSheetsService()
		if err != nil {
			return nil, err
		}

		resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, sheetName).MajorDimension("ROWS").Do()
		if err != nil {
			return nil, err
		}

		globalSheetData = &sheets.ValueRange{
			Values: resp.Values,
		}
	}

	var existingTracks []SpotifyTrack
	for _, row := range globalSheetData.Values {
		// Ensure the row has enough columns to represent a SpotifyTrack
		if len(row) < 4 {
			log.Warnf("Skipping row with insufficient columns: %v", row)
			continue
		}

		// Parse each column from the row
		name, nameOk := row[0].(string)
		author, authorOk := row[1].(string)
		url, urlOk := row[2].(string)
		playedAtStr, playedAtOk := row[3].(string)

		// Ensure the parsed data is valid
		if !nameOk || !authorOk || !urlOk || !playedAtOk {
			log.Warnf("Skipping row with invalid data: %v", row)
			continue
		}

		// Parse the playedAt string into a time.Time
		playedAt, err := time.Parse(timeLayout, playedAtStr)
		if err != nil {
			log.Warnf("Error parsing playedAt: %v", err)
			continue
		}

		// Create a SpotifyTrack struct and add it to the existingTracks slice
		track := SpotifyTrack{
			Name:     name,
			Author:   author,
			URL:      url,
			PlayedAt: playedAt,
		}
		existingTracks = append(existingTracks, track)
	}

	return existingTracks, nil
}
