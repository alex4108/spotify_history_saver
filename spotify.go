package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func authorize(ctx context.Context) *oauth2.Token {
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":"+os.Getenv("HTTP_PORT"), nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// Wait for auth to complete
	client := <-ch

	// Save token to file
	clientToken, err := client.Token()
	if err != nil {
		log.Errorf("Failed to fetch Spotify token: %v", err)
	}

	err = saveToken(clientToken)
	if err != nil {
		log.Errorf("Failed to save Spotify token: %v", err)
	}

	fmt.Println("Login Completed!")
	return clientToken
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// Use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	ch <- client

	// Save token to file
	saveToken(tok)

	fmt.Fprintf(w, "Login Completed!")
}

func loadToken() (*oauth2.Token, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		log.Warnf("Error opening spotify token file: %v", err)
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&token)
	if err != nil {
		log.Warnf("Error loading spotify token: %v", err)
		return nil, err
	}

	return &token, nil
}

// Function to save the Spotify token to a file
func saveToken(token *oauth2.Token) error {
	file, err := os.Create(tokenFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Save the latest token information
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(token)
	if err != nil {
		return err
	}

	return nil
}

func getRecentTracks(client *spotify.Client, userID string, limit int) ([]SpotifyTrack, error) {
	// Fetch recent tracks with a specified limit
	history, err := (client).PlayerRecentlyPlayedOpt(context.Background(), &spotify.RecentlyPlayedOptions{
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	var tracks []SpotifyTrack
	for _, item := range history {
		track := SpotifyTrack{
			Name:     item.Track.Name,
			Author:   item.Track.Artists[0].Name,
			URL:      item.Track.ExternalURLs["spotify"],
			PlayedAt: item.PlayedAt,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
