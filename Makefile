migrate:
	go run database/migration/main.go

test:
	go clean -testcache
	go test -coverprofile=cover.out -race ./...
	go tool cover -html cover.out -o cover.html

docker-build:
	docker build -tag gost
	docker compose up -d

# Please choose between windows/ unix
# unix
# openssl req -x509 -newkey rsa:2048 -keyout keys/private.key -out keys/publickey.crt -days 365 -nodes -subj "/CN=localhost"
# windows: using openssl.exe
# "C:\Program Files\Git\usr\bin\openssl.exe" req -x509 -newkey rsa:2048 -keyout keys/private.key -out keys/publickey.crt -days 365 -nodes -subj "/CN=localhost"