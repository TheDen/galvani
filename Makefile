.DEFAULT_GOAL := build
BINARY_NAME=galvani

prettier-format:
	# https://github.com/prettier/prettier
	@printf "%s\n" "==== Running prettier format ====="
	prettier -w . --log-level=error

prettier-lint:
	@printf "%s\n" "==== Running prettier lint check ====="
	prettier -c . --log-level=error

mod:
	go mod tidy
	go mod vendor
	go mod verify

go-update:
	go list -mod=readonly -m -f '{{if not .Indirect}}{{if not .Main}}{{.Path}}{{end}}{{end}}' all | xargs go get -u
	$(MAKE) mod

gosec:
	# https://github.com/securego/gosec
	gosec -severity medium  ./...

golines-format:
	# https://github.com/segmentio/golines
	@printf "%s\n" "==== Run golines ====="
	golines --write-output --ignored-dirs=vendor .

go-staticcheck:
	# https://github.com/dominikh/go-tools
	staticcheck ./...

compile:
	mkdir -p Galvani.app/Contents/MacOS/ || true
	CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin go build -ldflags="-s -w" -v -o ./bin/${BINARY_NAME}-darwin-amd64 main.go
	CGO_ENABLED=1 GOARCH=arm64 GOOS=darwin go build -ldflags="-s -w" -v -o ./bin/${BINARY_NAME}-darwin-arm64 main.go
	lipo -create -output Galvani.app/Contents/MacOS/${BINARY_NAME} bin/${BINARY_NAME}-darwin-amd64 bin/${BINARY_NAME}-darwin-arm64

build: compile
	mkdir -p Galvani.app/Contents/Resources/ || true
	./generate_icons.sh

run: build
	Galvani.app/Contents/MacOS/galvani

release: build compile
	rm *.dmg || true
	zip -r Galvani.zip Galvani.app
	create-dmg Galvani.app/
