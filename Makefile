# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]
	
# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/web: runs the cmd/web application
.PHONY: run/web
run/web:
	@go run ./cmd/web
	
# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/web: builds the cmd/web application
build/web:
	@go build -ldflags='-s' -o=./bin/web ./cmd/web
	
# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## docker/build: Builds the docker image
.PHONY: docker/build
docker/build:
	docker build . -t trolly -f devops/docker/Dockerfile

## docker/run: Runs the docker image at port 4000
.PHONY: docker/run
docker/run:
	docker run --network="host" -p 4000:4000 -p 3306:3306 trolly

## docker/stop: Stops the docker image
.PHONY: docker/stop
docker/stop:
	docker stop $$(docker ps -q --filter ancestor=trolly)

## docker-compose/up: Starts the docker compose stack
.PHONY: docker-compose/up
docker-compose/up:
	docker-compose -f devops/docker/docker-compose.yml up

# ==================================================================================== #
# TESTING
# ==================================================================================== #

## test/unit: runs all unit test in the cmd/web folder
.PHONY: test/unit
test/unit:
	@go test -v ./cmd/web/...

## test/coverage: generates a coverage report of all tests
.PHONY: test/coverage
test/coverage:
	@go test -coverprofile=/tmp/profile.out ./cmd/web/...
	@go tool cover -func=/tmp/profile.out

## test/coverage/web: opens browsers with generated coverage report
.PHONY: test/coverage/web
test/coverage/web:
	@go test -covermode=count -coverprofile=/tmp/profile.out ./cmd/web/...
	@go tool cover -html=/tmp/profile.out

