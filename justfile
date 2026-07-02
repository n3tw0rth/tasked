set dotenv-load

module := "github.com/n3tw0rth/tasked"

# Build with OAuth credentials from .env embedded via ldflags
build:
    #!/usr/bin/env bash
    set -euo pipefail
    : "${TASKED_CLIENT_ID:?set TASKED_CLIENT_ID in .env}"
    : "${TASKED_CLIENT_SECRET:?set TASKED_CLIENT_SECRET in .env}"
    go build -trimpath \
        -ldflags "-s -w \
            -X {{ module }}/internal/auth.clientID=${TASKED_CLIENT_ID} \
            -X {{ module }}/internal/auth.clientSecret=${TASKED_CLIENT_SECRET}" \
        -o tasked .

# Build without embedded credentials (uses TASKED_CLIENT_ID/SECRET env vars at runtime)
build-dev:
    go build -o tasked .

# Build and install into GOBIN
install: build
    #!/usr/bin/env bash
    set -euo pipefail
    bin="$(go env GOBIN)"
    [ -n "$bin" ] || bin="$(go env GOPATH)/bin"
    install -m 0755 tasked "$bin/tasked"

test:
    go test ./...

clean:
    rm -f tasked
