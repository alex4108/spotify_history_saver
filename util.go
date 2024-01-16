package main

import (
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func matchExists(track SpotifyTrack) bool {
	if globalSheetData == nil {
		_, err := getExistingTracksFromSheet(os.Getenv("GOOGLE_SHEET_ID"), os.Getenv("GOOGLE_SHEET_NAME"))
		if err != nil {
			log.Fatal("Error fetching existing tracks from Google Sheet: ", err)
		}
	}

	for _, value := range globalSheetData.Values {
		if len(value) > 0 {
			playedAt, err := time.Parse(timeLayout, value[3].(string))
			if err != nil {
				log.Warnf("Error parsing playedAt: %v", err)
				continue
			}

			trackInSheet := &SpotifyTrack{
				Name:     value[0].(string),
				Author:   value[1].(string),
				URL:      value[2].(string),
				PlayedAt: playedAt,
			}

			nameMatch := false
			if track.Name == trackInSheet.Name {
				log.Debugf("Found name match")
				nameMatch = true
			}

			authorMatch := false
			if track.Author == trackInSheet.Author {
				log.Debugf("Found author match")
				authorMatch = true
			}

			playedAtMatch := false
			if track.PlayedAt.Equal(trackInSheet.PlayedAt) {
				log.Debugf("Found playedat match")
				playedAtMatch = true
			} else {
				log.Debugf("Track Time: %v", track.PlayedAt)
				log.Debugf("Track In Sheet Time: %v", trackInSheet.PlayedAt)
			}

			trackMatch := false
			if track.URL == trackInSheet.URL {
				log.Debugf("Found track match")
				trackMatch = true
			}

			if trackMatch && playedAtMatch && authorMatch && nameMatch {
				return true
			}
		}
	}

	return false

}

func setLogLevelFromEnv() {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn", "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.Warnf("Invalid LOG_LEVEL '%s'. Defaulting to INFO level.", logLevel)
		log.SetLevel(log.InfoLevel)
	}
}
