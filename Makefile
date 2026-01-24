.DEFAULT_GOAL := build

.PHONY:fmt vet build run clean

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	cd cmd && go build -o ../bin && cd ..

run:
	./bin/cmd

clean:
	cd bin && rm ./* && cd ..
