test:
	go test -p 1 -timeout 240s -coverprofile=cover.out ./... 

test-race:
	go test -timeout 240s --race -coverprofile=cover.out ./...

test-per-dir:
	echo "./application"
	go test -p 1 -timeout 60s ./application/...
	echo "./controller"
	go test -p 1 -timeout 60s ./controller/...
	echo "./database"
	go test -p 1 -timeout 60s ./database/...
	echo "./domain"
	go test -p 1 -timeout 60s ./domain/...
	echo "./internal"
	go test -p 1 -timeout 60s ./internal/...
	echo "./repository"
	go test -p 1 -timeout 60s ./repository/...
	echo "./service"
	go test -p 1 -timeout 60s ./service/...
	echo "./tests"
	go test -p 1 -timeout 60s ./tests/...

migrate:
	go run database/migration/main.go

run:
	go run .

test-report:
	go tool cover -html cover.out -o cover.html

migrate-test-report:
	go run database/migration/main.go
	timeout 5
	go test -p 1 -timeout 240s -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html

# !!windows only!! run redis in background
st-redis:
	redis-server.exe --service-start