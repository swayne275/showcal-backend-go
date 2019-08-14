# Convenience Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
BINARY_NAME=showcal-backend
BINARY_UNIX=$(BINARY_NAME)_unix

target: ;

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
test:
	$(GOTEST) -v -race ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GOGET) "github.com/swayne275/gerrors"
	$(GOGET) "github.com/tidwall/gjson"
	$(GOGET) "golang.org/x/net/context"
	$(GOGET) "golang.org/x/oauth2"
	$(GOGET) "golang.org/x/oauth2/google"
	$(GOGET) "google.golang.org/api/calendar/v3"
	$(GOINSTALL) "github.com/swayne275/showcal-backend-go/clientapi
	$(GOINSTALL) "github.com/swayne275/showcal-backend-go/gcalwrapper
	$(GOINSTALL) "github.com/swayne275/showcal-backend-go/tvshowdata

# Cross compilation (not needed yet)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -0 $(BINARY_UNIX) -v
# add docker build down the road when needed from:
# https://sohlich.github.io/post/go_makefile/