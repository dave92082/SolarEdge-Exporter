build:
	mkdir -p ./bin/linux
	mkdir -p ./bin/windows
	mkdir -p ./bin/darwin
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/solaredge-exporter main.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/windows/solaredge-exporter.exe main.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/darwin/solaredge-exporter main.go
	zip ./bin/solaredge-exporter-linux-1.0.0.zip ./bin/linux/solaredge-exporter
	zip ./bin/solaredge-exporter-windows-1.0.0.zip ./bin/windows/solaredge-exporter.exe
	zip ./bin/solaredge-exporter-macos-1.0.0.zip ./bin/darwin/solaredge-exporter