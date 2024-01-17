# Spotify History Saver

[![Build and Push](https://github.com/alex4108/spotify_history_saver/actions/workflows/build.yml/badge.svg)](https://github.com/alex4108/spotify_history_saver/actions/workflows/build.yml)
[![Docker Image Version](https://img.shields.io/docker/v/alex4108/spotify_history_saver)](https://hub.docker.com/alex4108/spotify_history_saver)
[![GitHub issues](https://img.shields.io/github/issues/alex4108/spotify_history_saver)](https://github.com/alex4108/spotify_history_saver/issues)
[![Docker Pulls](https://img.shields.io/docker/pulls/alex4108/spotify_history_saver)](https://hub.docker.com/alex4108/spotify_history_saver)

[Latest Releases](https://github.com/alex4108/spotify_history_saver/releases)

## What does it do?

This container pulls your recent spotify play history using the [Spotify API](https://developer.spotify.com/documentation/web-api/reference/get-recently-played).  It then records this data in a sorted, deduplicated manner in [Google Sheets](https://sheets.google.com)

I use this so I can persist the data about my spotify plays and do my own analysis on it in the future.

If you found this project helpful please Star the repo!

## Usage

This app is published as a docker container.  The `:latest` tag *should* be stable, although recommend Watching the repo and pinning build tags.

## Configuration / Environment Variables

Take a browse through example.env for all the configuration required.

### Google Cloud 

* Spin up a project in [Google Cloud Console](https://cloud.google.com).
* Give the project Google Sheets API Access from APIs & Services button.
* From Credentials on the left, create a Service Account credential.
* Once created, a .json file will be downloaded
* Encode the JSON file to be injected to `GOOGLE_SHEETS_CREDENTIAL`: `cat google-sa-creds.json | base64 | tr -d '\n'`
* Navigate to the Google Sheet you wish to use, and copy the ID from the URL.  Save this as `GOOGLE_SHEET_ID`
* Share the document to the new Service Account's email, give it Editor permissions.

### Spotify Developer

* Generate a Client ID & Secret with Web API access from [Spotify Developer Portal](https://developer.spotify.com/dashboard/create).  
* Set the Redirect URI to `http://localhost:8080/callback` if you use the default HTTP_PORT value.  Set this to a more appropriate host if you've internet exposed your deployment.
* Ensure you set your spotify account's email address in Settings > User Management.
* Grab your `SPOTIFY_USER_ID` following [these instructions](https://www.bonjohh.com/how-to-get-my-spotify-user-id.html).

_Credentials are persisted to the path defined in `SPOTIFY_TOKEN_FILE` so as to not require login on each run!_

## Metrics

Prometheus scraper endpoint is available on `http://app_server:8081/metrics`

## Daemon Mode

If you intend to run this as a constantly-running service, set `DAEMON=1` and `DAEMON_SLEEP_SECS` to the frequency to run at, eg 60 for 1 minute execution frequency.