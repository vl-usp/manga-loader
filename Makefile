MANGA_NAME = 8kaijuu
VOL_NUM = 2
WORKERS = 16

run: build
	./bin/loader -name=$(MANGA_NAME) -vol_num=$(VOL_NUM) -workers=$(WORKERS)

build:
	mkdir -p bin
	mkdir -p output
	go build -o ./bin/loader main.go