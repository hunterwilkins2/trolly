MAIN_PACKAGE_PATH := ./cmd/web
BINARY_NAME := trolly

## run: runs this package with hot reloading when saved
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0

## tailwind/build: complies tailwind css
.PHONY: tailwind/build
tailwind/build:
	tailwindcss -i static/css/input.css -o static/css/dist/output.css --minify

## templ/build: compiles templ files
.PHONY: templ/build
templ/build:
	templ generate

## build: builds this package
.PHONY: build
build: tailwind/build templ/build
	go build -o=bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## db: starts a MySQL docker container
.PHONY: db
db:
	docker run --name trolly-db -p 3306:3306 -e MARIADB_DATABASE=trolly -e MARIADB_ROOT_PASSWORD=admin -e MARIADB_USER=trolly -e MARIADB_PASSWORD=pa55word -d mariadb:latest

migrate:
	migrate -path=./migrations -database="mysql://trolly:pa55word@tcp(0.0.0.0:3306)/trolly" up

# Utilites
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'