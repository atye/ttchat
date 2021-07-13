build:
	-mkdir -p ./bin
	CGO_ENABLED=0 go build -o ./bin/ttchat ./main.go

test:
	go test -v -count=1 -v -cover -race ./...

cover-profile:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm -f coverage.out