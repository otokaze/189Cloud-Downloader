# Binary name
BINARY=189Cloud-Downloader
VERSION=$(shell grep -E -o  v[0-9]+\.[0-9]+\.[0-9]+ CHANGELOG.md | head -1)
BUILD_FLAGS="-s -w -X main.version=${VERSION}"

# Builds the project
build:
		go build -o ${BINARY} -ldflags ${BUILD_FLAGS}
# Installs our project: copies binaries
install:
		go install
release:
		# Clean	
		go clean
		rm -rf *.gz ${BINARY}-*
		# Build for darwin amd64
		$(eval FILENAME=${BINARY}-darwin-amd64-${VERSION})
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${FILENAME} -ldflags ${BUILD_FLAGS}
		tar czvf ${FILENAME}.tar.gz ./${FILENAME}
		# Build for linux arm64
		$(eval FILENAME=${BINARY}-linux-arm64-${VERSION})
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${FILENAME} -ldflags ${BUILD_FLAGS}
		tar czvf ${FILENAME}.tar.gz ./${FILENAME}
		# Build for linux amd64
		$(eval FILENAME=${BINARY}-linux-amd64-${VERSION})
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${FILENAME} -ldflags ${BUILD_FLAGS}
		tar czvf ${FILENAME}.tar.gz ./${FILENAME}
		# Build for windows amd64
		$(eval FILENAME=${BINARY}-windows-amd64-${VERSION}.exe)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64  go build -o ${FILENAME} -ldflags ${BUILD_FLAGS}
		tar czvf ${FILENAME}.tar.gz ./${FILENAME}
# Cleans our projects: deletes binaries
clean:
		go clean
		rm -rf *.gz ${BINARY}-*

.PHONY:  clean build