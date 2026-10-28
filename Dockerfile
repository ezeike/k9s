# -----------------------------------------------------------------------------
# The base image for building the k9s binary
FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
RUN apk --no-cache add --update make libx11-dev git gcc libc-dev curl \
  && make build

# -----------------------------------------------------------------------------
# Build the final Docker image for the target platform (not the build host).
# Pinning this stage to $BUILDPLATFORM would yield an amd64 runtime image (incl.
# kubectl) even for the arm64 manifest entry, so we let it default to
# $TARGETPLATFORM and use buildx's TARGETARCH to fetch the matching kubectl.
FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
ARG KUBECTL_VERSION="v1.35.3"
ARG TARGETARCH

COPY --from=build /k9s/execs/k9s /bin/k9s
RUN apk --no-cache add --update ca-certificates \
  && apk --no-cache add --update -t deps curl vim \
  && curl -f -L https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && apk del --purge deps

ENTRYPOINT [ "/bin/k9s" ]
