package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	tokenFile = ""
	auth      = spotifyauth.New()
	ch        = make(chan *spotify.Client)
	state     = "abc123"
)

const (
	timeLayout = "2006-01-02T15:04:05.999999999Z07:00" // RFC3339 format with nanoseconds
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn(".env file not found!")
	}
	setLogLevelFromEnv()

	auth = spotifyauth.New(
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		spotifyauth.WithRedirectURL(os.Getenv("SPOTIFY_REDIRECT_URI")),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadRecentlyPlayed),
	)

	ctx := context.TODO()

	tokenFile = os.Getenv("SPOTIFY_TOKEN_FILE")

	token, err := loadToken()
	if err != nil {
		token = authorize(ctx)
	}

	client := spotify.New(auth.Client(ctx, token))

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("You are logged in as: %v", user.ID)

	// Save token to file
	clientToken, err := client.Token()
	if err != nil {
		log.Errorf("Failed to fetch Spotify token: %v", err)
	}

	err = saveToken(clientToken)
	if err != nil {
		log.Errorf("Failed to save Spotify token during init: %v", err)
	}

	// Get recent tracks from Spotify with a limit of 50 songs
	songLimit := 50 // Maximum
	recentTracks, err := getRecentTracks(client, os.Getenv("SPOTIFY_USER_ID"), songLimit)
	if err != nil {
		log.Fatal("Error getting recent tracks from Spotify: ", err)
	}

	// Write the data to Google Sheet
	writeToSheet(os.Getenv("GOOGLE_SHEET_ID"), os.Getenv("GOOGLE_SHEET_NAME"), recentTracks)
	log.Infof("Spotify history recorded successfully.")
}
