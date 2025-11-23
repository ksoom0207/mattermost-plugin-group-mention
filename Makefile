.PHONY: all build test clean dist check-style webapp server

GO_BUILD_FLAGS = -trimpath
GO_TEST_FLAGS = -v -race
PLUGIN_ID = com.mattermost.plugin-group-mention
PLUGIN_VERSION = 1.0.0
BUNDLE_NAME = $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz

all: check-style test dist

## Builds the server and webapp.
build: server webapp

## Builds the server.
server:
	mkdir -p server/dist
	cd server && go build $(GO_BUILD_FLAGS) -o dist/plugin-linux-amd64 ./main.go

## Builds the webapp.
webapp:
	cd webapp && npm install && npm run build

## Runs Go tests.
test:
	cd server && go test $(GO_TEST_FLAGS) ./...

## Checks code style.
check-style:
	cd server && go vet ./...
	cd server && gofmt -s -d .

## Creates the plugin bundle.
dist: build
	rm -rf dist
	mkdir -p dist/$(PLUGIN_ID)
	cp plugin.json dist/$(PLUGIN_ID)/
	mkdir -p dist/$(PLUGIN_ID)/server/dist
	cp -r server/dist/* dist/$(PLUGIN_ID)/server/dist/
	mkdir -p dist/$(PLUGIN_ID)/webapp/dist
	cp -r webapp/dist/* dist/$(PLUGIN_ID)/webapp/dist/
	cd dist && tar -cvzf $(BUNDLE_NAME) $(PLUGIN_ID)
	@echo Plugin built at: dist/$(BUNDLE_NAME)

## Cleans build artifacts.
clean:
	rm -rf server/dist
	rm -rf webapp/dist
	rm -rf webapp/node_modules
	rm -rf dist
	rm -rf build

## Deploys the plugin to a local Mattermost instance (set MM_SERVICESETTINGS_SITEURL).
deploy: dist
	curl -F "plugin=@dist/$(BUNDLE_NAME)" \
		$(MM_SERVICESETTINGS_SITEURL)/api/v4/plugins \
		-H "Authorization: Bearer $(MM_ADMIN_TOKEN)"
