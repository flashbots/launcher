default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -o ./bin/launcher github.com/flashbots/launcher/cmd

.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: release
release:
	@rm -rf ./dist
	GITHUB_TOKEN=$$( gh auth token ) goreleaser release
