test:
	go test -timeout 60s -coverprofile=cover.out ./...

migrate:
	go run database/migration/main.go

run:
	go run .

test-report:
	go tool cover -html cover.out -o cover.html