migrate:
	go run database/migration/main.go

run:
	go run .

test:
	go clean -testcache
	go test -coverprofile=cover.out -p 6 ./...
	go tool cover -html cover.out -o cover.html

migrate-test-report:
	go run database/migration/main.go
	timeout 5
	go clean -testcache
	go test -coverprofile=cover.out -race ./...
	go tool cover -html cover.out -o cover.html

docker-build:
	docker build -tag gost
	timeout 5
	docker compose up -d