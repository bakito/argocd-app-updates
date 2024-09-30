# Include toolbox tasks
include ./.toolbox.mk

# Run go golanci-lint
lint: golangci-lint
	$(LOCALBIN)/golangci-lint run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: tidy lint
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

release: semver
	@version=$$($(LOCALBIN)/semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --clean

test-release:
	goreleaser --skip-publish --snapshot --clean

