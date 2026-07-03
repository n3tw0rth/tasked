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

# Build and install (default ~/.local/bin; override: just install /usr/local/bin)
install dest="~/.local/bin": build
    #!/usr/bin/env bash
    set -euo pipefail
    dest="{{ dest }}"
    dest="${dest/#\~/$HOME}"
    mkdir -p "$dest"
    install -m 0755 tasked "$dest/tasked"
    echo "installed $dest/tasked"
    case ":$PATH:" in
        *":$dest:"*) ;;
        *) echo "warning: $dest is not on your PATH" ;;
    esac

# Install shell completions for the current shell (bash|zsh|fish)
completions shell="bash":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{ shell }}" in
        bash)
            dir="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions"
            mkdir -p "$dir"
            ./tasked completion bash > "$dir/tasked" ;;
        zsh)
            dir="${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions"
            mkdir -p "$dir"
            ./tasked completion zsh > "$dir/_tasked"
            echo "ensure $dir is in your zsh fpath" ;;
        fish)
            dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions"
            mkdir -p "$dir"
            ./tasked completion fish > "$dir/tasked.fish" ;;
        *)
            echo "unsupported shell: {{ shell }} (use bash, zsh, or fish)" >&2
            exit 1 ;;
    esac
    echo "installed {{ shell }} completions"

test:
    go test ./...

clean:
    rm -f tasked
