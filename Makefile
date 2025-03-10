.PHONY: all
all: build


.PHONY: test
test: run-unit-tests run-e2e-tests-local

.PHONY: run-unit-tests
run-unit-tests:
	docker build . -t gotest && docker run -t --rm --name draft-test gotest test ./... -buildvcs=false

#TODO: add more e2e tests to the local testing
.PHONY: run-e2e-tests-local
run-e2e-tests-local: build
	test/check_info_schema.sh;

.PHONY: build
build:
	GO111MODULE=on go build -v -o .

.PHONY: build-all
build-all: build-windows-amd64 build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/Azure/draft/cmd.VERSION=${DRAFT_VERSION}" -v -o ./bin/draft-windows-amd64.exe

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/Azure/draft/cmd.VERSION=${DRAFT_VERSION}" -v -o ./bin/draft-linux-amd64

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/Azure/draft/cmd.VERSION=${DRAFT_VERSION}" -v -o ./bin/draft-linux-arm64

.PHONY: build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/Azure/draft/cmd.VERSION=${DRAFT_VERSION}" -v -o ./bin/draft-darwin-amd64

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/Azure/draft/cmd.VERSION=${DRAFT_VERSION}" -v -o ./bin/draft-darwin-arm64

.PHONY: clean-entra-app
clean-entra-app:
	@read -p "Enter the display name of the Azure entra application to delete: " APP_DISPLAY_NAME; \
	APP_ID_TO_DELETE=$$(az ad app list --display-name $$APP_DISPLAY_NAME | jq -r '.[].appId'); \
	if [ -z "$$APP_ID_TO_DELETE" ]; then \
	  echo "No Azure entra application found with the specified display name: $$APP_DISPLAY_NAME"; \
	else \
	  az ad app delete --id $$APP_ID_TO_DELETE; \
	  echo "Deleted Azure entra application with display name: $$APP_DISPLAY_NAME"; \
	fi
