.PHONY: dist

PLATFORMS = linux darwin
ARCHITECTURES = amd64 arm64

dist:
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			GOOS=$$platform GOARCH=$$arch go build -trimpath -o dist/unpath-$$platform-$$arch main.go; \
		done \
	done
