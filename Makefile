VOL_FROM = 1
VOL_TO = 1
WORKERS = 8
DATA_FOLDER_NAME = ./data

run: build
	./bin/downloader -json=$(DATA_FOLDER_NAME)/manga.json -from=$(VOL_FROM) -to=$(VOL_TO) -workers=$(WORKERS)

build:
	go build -C downloader -o ../bin/downloader main.go