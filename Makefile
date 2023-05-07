.DEFAULT_GOAL := build
BINARY_NAME=galvani

compile:
	mkdir -p Galvani.app/Contents/MacOS/ || true
	CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin go build -ldflags="-s -w" -v -o ./bin/${BINARY_NAME}-darwin-amd64 main.go
	CGO_ENABLED=1 GOARCH=arm64 GOOS=darwin go build -ldflags="-s -w" -v -o ./bin/${BINARY_NAME}-darwin-arm64 main.go
	lipo -create -output Galvani.app/Contents/MacOS/${BINARY_NAME} bin/${BINARY_NAME}-darwin-amd64 bin/${BINARY_NAME}-darwin-arm64

build: compile
	./generate_icons.sh

run: build
	Galvani.app/Contents/MacOS/galvani

release: build compile
	rm *.dmg || true
	zip -r Galvani.zip Galvani.app
	create-dmg Galvani.app/
