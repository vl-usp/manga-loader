MANGA_NAME = 8kaijuu
VOL_NUM = 2
WORKERS = 8
DEBUG = false

run: build
	./bin/loader -name=$(MANGA_NAME) -vol_num=$(VOL_NUM) -workers=$(WORKERS) -debug=$(DEBUG)

build:
	mkdir -p bin
	mkdir -p output
	go build -o ./bin/loader cmd/loader/main.go