all:
	env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -v
