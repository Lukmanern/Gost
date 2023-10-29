test:
	go test -timeout 200s -coverprofile=cover.out ./...

test-race:
	go test -timeout 200s --race -coverprofile=cover.out ./...

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

# !!windows only!! run redis in background
st-redis:
	redis-server.exe --service-start