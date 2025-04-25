# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

buildversion := $(shell git describe --tags --always --dirty)
buildtime := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

## build/rpi: build the application for Raspberry Pi using ARMHF (any version)
.PHONY: build/rpi
build/rpi:
	@docker build \
	--build-arg GOOS=linux --build-arg GOARCH=arm \
	--build-arg BUILDTIME=$(buildtime) --build-arg BUILDVERSION=$(buildversion) \
	--output=out --target=binaries-armhf --progress=plain \
	-f build/docker/Dockerfile .

## build/rpi/v6: build the application for Raspberry Pi ARMv6 (1*, Zero*)
.PHONY: build/rpi/v6
build/rpi/v6:
	@docker build \
	--build-arg GOOS=linux --build-arg GOARCH=arm --build-arg GOARM=6 \
	--build-arg BUILDTIME=$(buildtime) --build-arg BUILDVERSION=$(buildversion) \
	--output=out --target=binaries-armel --progress=plain \
	-f build/docker/Dockerfile .

## build/rpi/v8/64: build the application for Raspberry Pi ARMv8 64bit (3*, 4*, Zero 2W)
.PHONY: build/rpi/v8/64
build/rpi/v8/64:
	@docker build \
	--build-arg GOOS=linux --build-arg GOARCH=arm64 \
	--build-arg BUILDTIME=$(buildtime) --build-arg BUILDVERSION=$(buildversion) \
	--output=out --target=binaries-arm64-8 --progress=plain \
	-f build/docker/Dockerfile .