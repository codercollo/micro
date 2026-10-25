SHELL=cmd.exe
FRONT_END_BINARY=frontApp.exe
BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
LOGGER_BINARY=loggerServiceApp
MAIL_BINARY = mailerApp
LISTENER_BINARY= listenerApp
FRONT_BINARY= frontEndApp

## up: starts all containers in the background without forcing build
up:
	@echo Starting Docker images...
	docker-compose up -d
	@echo Docker images started!

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_broker build_auth build_logger build_mail build_listener
	@echo Stopping docker images (if running...)
	docker-compose down
	@echo Building (when required) and starting docker images...
	docker-compose up --build -d
	@echo Docker images built and started!

## down: stop docker compose
down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!

## build_front_end_linux: builds the front-end linux binary as a linux executable
build_front_linux:
	@echo Building front end linux binary...
	cd ..\front-end && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${FRONT_BINARY} ./cmd/web
	@echo Done!

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo Building broker binary...
	cd ..\broker-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${BROKER_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo Building logger binary...
	cd ..\logger-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!	

## build_mail: builds the mailer binary as a linux executable
build_mail:
	@echo Building logger binary...
	cd ..\mail-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${MAIL_BINARY} ./cmd/api
	@echo Done!	

## build_auth: builds the broker binary as a linux executable
build_auth:
	@echo Building auth binary...
	cd ..\authentication-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${AUTH_BINARY} ./cmd/api
	@echo Done!	

## build_listener: builds the broker binary as a linux executable
build_listener:
	@echo Building listener binary...
	cd ..\listener-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o ${LISTENER_BINARY} .
	@echo Done!		

## build_front: builds the front end binary
build_front:
	@echo Building front end binary...
	cd ..\front-end && set CGO_ENABLED=0&& set GOOS=windows&& go build -o ${FRONT_END_BINARY} ./cmd
	@echo Done!

## start: starts the front end
start: build_front
	@echo Starting front end
	cd ..\front-end && start /B ${FRONT_END_BINARY}

## stop: stop the front end
stop:
	@echo Stopping front end...
	@taskkill /IM "${FRONT_END_BINARY}" /F
	@echo Stopped front end!