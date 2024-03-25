VERSION := "$(shell git describe --tags)-$(shell git rev-parse --short HEAD)"
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X github.com/stefan-chivu/go-template/go-template.Version=$(VERSION)
GOLDFLAGS += -X github.com/stefan-chivu/go-template/go-template.Buildtime=$(BUILDTIME)

GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build release

build: clean
	go build -o go-template-app $(GOFLAGS) ./go-template
	chmod +x go-template-app
	./go-template-app -version

clean:
	rm -f go-template-app
	rm -f cover.out
	rm -f cpu.pprof

cover:
	go test -count=1 -cover -coverprofile=cover.out ./...
	go tool cover -func=cover.out

debug: build
	./go-template-app -PProf -CPUProfile=cpu.pprof -ServerTLSCert=server.crt -ServerTLSKey=server.key

lint:
	go fmt ./ ./go-template/...
	go vet

release:
	mkdir -p release
	rm -f release/go-template-app release/go-template-app.exe
ifeq ($(shell go env GOOS), windows)
	go build -o release/go-template-app.exe $(GOFLAGS) ./go-template
	cd release; zip -m "go-template-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" go-template-app.exe
else
	go build -o release/go-template-app $(GOFLAGS) ./go-template
	cd release; zip -m "go-template-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" go-template-app
endif
	cd release; sha256sum "go-template-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" > "go-template-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip.sha256"


run: build
	./go-template-app -ServerTLSCert=server.crt -ServerTLSKey=server.key

sync:
	go get ./...

test: clean
	go test -count=1 -cover ./...

tls:
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj "/CN=selfsigned.go-template.local"

update:
	go mod tidy
	go get -u ./...

# build-client-deps: clean-client
# 	npm whatever

run-client: # build-client
	# ./go-template-webclient
	npm --prefix `pwd`/webclient start 
