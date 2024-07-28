VOL_FROM = 1
VOL_TO = 1
WORKERS = 8
DEBUG = false

run: build
	./bin/downloader -json=./data/$(MANGA_NAME)/manga.json -from=$(VOL_FROM) -to=$(VOL_TO) -workers=$(WORKERS) -debug=$(DEBUG)

build:
	go build -C downloader -o ../bin/downloader main.go