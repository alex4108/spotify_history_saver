.PHONY: all build build-container clean

all: build build-container

build:
	go build -o spotify-history-saver .

run: build
	sed -i 's/DAEMON=.*/DAEMON=0/g' .env
	./spotify-history-saver

run-daemon: build
	sed -i 's/DAEMON=.*/DAEMON=1/g' .env
	./spotify-history-saver

build-container:
	docker build -t alex4108/spotify_history_saver:latest .

push-container: build-container
	docker push alex4108/spotify_history_saver:latest

clean:
	rm -f main
	