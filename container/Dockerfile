FROM cf-registry.cyberjake.xyz/library/golang:1.24-alpine AS builder
# Set destination for COPY
WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apk,sharing=locked \
	--mount=type=cache,target=/var/lib/apk,sharing=locked \
	apk update && apk upgrade && \
    apk add git tzdata ca-certificates

# RUN apk update --no-cache && apk upgrade --no-cache \
#     && apk add --no-cache git tzdata ca-certificates

# Download Go modules
COPY ./container/go.mod ./container/go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

# Copy container src
COPY container/*.go ./
COPY .git ./
# Build
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -ldflags "-X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o /server


FROM cf-registry.cyberjake.xyz/library/alpine:latest
COPY --from=builder /server /server
EXPOSE 8080

CMD ["/server"]
