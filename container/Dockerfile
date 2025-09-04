FROM cf-registry.cyberjake.xyz/library/golang:1.24-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder
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
COPY ./container/*.go ./
COPY .git ./
# Build
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -ldflags "-X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o /server


FROM library/alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
COPY --from=builder /server /server
EXPOSE 8080

CMD ["/server"]
