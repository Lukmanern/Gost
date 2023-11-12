migrate:
	go run database/migration/main.go

run:
	go run .

test-clear:
	go clean -testcache
	
test:
	go clean -testcache
	go test -p 1 -timeout 330s -coverprofile=cover.out ./... 
	go tool cover -html cover.out -o cover.html
	del cover.out

test-race:
	go clean -testcache
	go test -race -timeout 200s ./controller/... ./database/...  ./internal/... ./repository/... ./service/...

migrate-test-report:
	go run database/migration/main.go
	timeout 5
	go clean -testcache	
	go test -race -timeout 200s ./controller/... ./database/...  ./internal/... ./repository/... ./service/...
	go test -p 1 -timeout 330s -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html
	del cover.out

# windowsOS only
st-redis:
	redis-server.exe --service-start

# unix only
generate-keys:
	openssl req -x509 -newkey rsa:4096 -keyout keys/private.key -out keys/publickey.crt -days 365 -nodes -subj "/CN=localhost"