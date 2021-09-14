APP=wptest
BIN_FOLDER=bin

.SiLENT:

lint:
#	golangci-lint run -c ./.golangci.yml > lint.txt
	golangci-lint run
.PHONY:lint


run:
	go run cmd/${APP}/main.go
.PHONY:run

##
## Building
##
install: 
	go mod download

build: 
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ${BIN_FOLDER}/app $(shell go list -m)/cmd/${APP}

.PHONY: install build