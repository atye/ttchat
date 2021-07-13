build:
	go build -o bin/ttchat cmd/main.go

test:
	go test -v -count=1 -v -cover -race ./...

cover-profile:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm -f coverage.out