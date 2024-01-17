package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	tokenFile = ""
	auth      = spotifyauth.New()
	ch        = make(chan *spotify.Client)
	state     = "abc123"

	// Define Prometheus metrics
	spotifyRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "spotify_request_duration_seconds",
			Help: "Histogram of the duration of Spotify API requests.",
		},
		[]string{"endpoint"},
	)

	googleRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "googlesheets_request_duration_seconds",
			Help: "Histogram of the duration of Google Sheets API requests.",
		},
		[]string{"endpoint"},
	)

	authenticated = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shs_authenticated",
			Help: "1 if SHS is authenticated, 0 if waiting for auth",
		},
		[]string{},
	)

	lastRunSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shs_success",
			Help: "1 if last run was clean, 0 if not",
		},
		[]string{},
	)
)

const (
	timeLayout = "2006-01-02T15:04:05.999999999Z07:00" // RFC3339 format with nanoseconds
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(spotifyRequestDuration)
	prometheus.MustRegister(googleRequestDuration)
	prometheus.MustRegister(authenticated)
	prometheus.MustRegister(lastRunSuccess)
}

func main() {
	err := godotenv.Load()
	setLogLevelFromEnv()
	if err != nil {
		log.Info(".env file not found!")
	}

	// Expose Prometheus metrics endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":"+os.Getenv("METRICS_HTTP_PORT"), nil)
		if err != nil {
			log.Fatal("Failed to start Prometheus metrics server: ", err)
		}
	}()

	if os.Getenv("DAEMON") == "1" {
		seconds, err := strconv.ParseInt(os.Getenv("DAEMON_SLEEP_SECS"), 10, 0)
		if err != nil {
			log.Fatal("Invalid DAEMON_SLEEP_SECS: %v", err)
		}
		run()

		ticker := time.NewTicker(time.Duration(seconds) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				log.Debugf("Woke up after %v seconds", os.Getenv("DAEMON_SLEEP_SECS"))
				run()
			}
		}()

		select {}
	} else {
		run()
	}
}

func run() {
	lastRunSuccess.WithLabelValues().Set(0)

	auth = spotifyauth.New(
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		spotifyauth.WithRedirectURL(os.Getenv("SPOTIFY_REDIRECT_URI")),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadRecentlyPlayed),
	)

	ctx := context.TODO()

	tokenFile = os.Getenv("SPOTIFY_TOKEN_FILE")

	authenticated.WithLabelValues().Set(0)
	token, err := loadToken()
	if err != nil {
		token = authorize(ctx)
	}

	client := spotify.New(auth.Client(ctx, token))

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	authenticated.WithLabelValues().Set(1)
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
	startTime := time.Now()
	recentTracks, err := getRecentTracks(client, os.Getenv("SPOTIFY_USER_ID"), songLimit)
	duration := time.Since(startTime).Seconds()

	if err != nil {
		log.Fatal("Error getting recent tracks from Spotify: ", err)
	}

	// Record the duration of the Spotify API request
	spotifyRequestDuration.WithLabelValues("getRecentTracks").Observe(duration)

	// Write the data to Google Sheet
	startTime = time.Now()
	writeToSheet(os.Getenv("GOOGLE_SHEET_ID"), os.Getenv("GOOGLE_SHEET_NAME"), recentTracks)
	duration = time.Since(startTime).Seconds()
	googleRequestDuration.WithLabelValues("writeToSheet").Observe(duration)
	lastRunSuccess.WithLabelValues().Set(1)
	log.Infof("Spotify history recorded successfully.")
}
