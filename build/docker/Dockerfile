# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-bullseye AS base
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x



FROM base AS base-armhf
RUN dpkg --add-architecture armhf
RUN apt-get update && apt-get install -y libasound2-dev:armhf gcc-arm-linux-gnueabihf

FROM base AS base-armel
RUN dpkg --add-architecture armel
RUN apt-get update && apt-get install -y libasound2-dev:armel gcc-arm-linux-gnueabi

FROM base AS base-arm64
RUN dpkg --add-architecture arm64
RUN apt-get update && apt-get install -y libasound2-dev:arm64 gcc:arm64



FROM base-armhf as build-app-armhf
WORKDIR /src
ARG CGO_ENABLED=1
ARG CC=arm-linux-gnueabihf-gcc
ARG PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabihf/pkgconfig
ARG GOOS
ARG BUILDVERSION
ARG BUILDTIME
RUN --mount=type=bind,target=. find /src -type f
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=${GOOS} GOARCH=arm \
    go build -ldflags "-X 'main.Version=${BUILDVERSION}' -X main.build=${BUILDTIME}" -o /out/homekit-garage-shutter-${GOOS}-armhf

FROM base-armel as build-app-armel
WORKDIR /src
ARG CGO_ENABLED=1
ARG CC=arm-linux-gnueabi-gcc
ARG PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabi/pkgconfig
ARG GOOS
ARG GOARM
ARG BUILDVERSION
ARG BUILDTIME
RUN --mount=type=bind,target=. find /src -type f
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=${GOOS} GOARCH=arm GOARM=${GOARM} \
    go build -ldflags "-X 'main.Version=${BUILDVERSION}' -X main.build=${BUILDTIME}" -o /out/homekit-garage-shutter-${GOOS}-armel-${GOARM}

FROM base-armel as build-app-armel-8
WORKDIR /src
ARG CGO_ENABLED=1
ARG CC=arm-linux-gnueabi-gcc
ARG PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabi/pkgconfig
ARG GOOS
ARG BUILDVERSION
ARG BUILDTIME
RUN --mount=type=bind,target=. find /src -type f
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=${GOOS} GOARCH=arm \
    go build -ldflags "-X 'main.Version=${BUILDVERSION}' -X main.build=${BUILDTIME}" -o /out/homekit-garage-shutter-${GOOS}-armel-8

FROM base-arm64 as build-app-arm64-8
WORKDIR /src
ARG CGO_ENABLED=1
ARG GOOS
ARG BUILDVERSION
ARG BUILDTIME
RUN --mount=type=bind,target=. find /src -type f
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=${GOOS} GOARCH=arm64 \
    go build -ldflags "-X 'main.Version=${BUILDVERSION}' -X main.build=${BUILDTIME}" -o /out/homekit-garage-shutter-${GOOS}-arm64-8



FROM scratch AS binaries-armhf
ARG GOOS
ARG GOARM
COPY --from=build-app-armhf /out/*-${GOOS}-armhf /

FROM scratch AS binaries-armel
ARG GOOS
ARG GOARM
COPY --from=build-app-armel /out/*-${GOOS}-armel-${GOARM} /

FROM scratch AS binaries-armel-8
ARG GOOS
COPY --from=build-app-armel-8 /out/*-${GOOS}-armel-8 /

FROM scratch AS binaries-arm64-8
ARG GOOS
COPY --from=build-app-arm64-8 /out/*-${GOOS}-arm64-8 /
