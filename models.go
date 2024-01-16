package main

import "time"

type SpotifyTrack struct {
	PlayedAt time.Time
	Name     string
	Author   string
	URL      string
}
