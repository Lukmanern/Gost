test:
	go test -timeout 200s -coverprofile=cover.out ./...

migrate:
	go run database/migration/main.go

run:
	go run .

test-report:
	go tool cover -html cover.out -o cover.html

migrate-test-report:
	go run database/migration/main.go
	timeout 5
	go test -timeout 200s -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html