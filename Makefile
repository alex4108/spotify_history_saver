.PHONY: all build build-container clean

all: build build-container

build:
	go build -o spotify-history-saver .

run: build
	./spotify-history-saver

build-container:
	docker build -t alex4108/spotify_history_saver:latest .

push-container: build-container
	docker push alex4108/spotify_history_saver:latest

clean:
	rm -f main
	