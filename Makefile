migrate:
	go run database/migration/main.go

run:
	go run .

test-clear:
	go clean -testcache
	
test:
	go clean -testcache
	go test -p 1 -timeout 240s -coverprofile=cover.out ./... 

test-race:
	go clean -testcache
	go test -race -timeout 200s ./controller/... ./database/...  ./internal/... ./repository/... ./service/...

test-report:
	go tool cover -html cover.out -o cover.html

migrate-test-report:
	go clean -testcache
	go run database/migration/main.go
	timeout 5
	go test -p 1 -timeout 240s -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html

# windowsOS only
st-redis:
	redis-server.exe --service-start