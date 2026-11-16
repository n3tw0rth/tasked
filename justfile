module := "github.com/n3tw0rth/tasked"
version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

# Build the binary
build:
    go build -trimpath -ldflags "-s -w -X {{ module }}/cmd.version={{ version }}" -o tasked .

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

# Everything CI runs: formatting, vet, tests
check:
    #!/usr/bin/env bash
    set -euo pipefail
    unformatted="$(gofmt -l .)"
    if [ -n "$unformatted" ]; then echo "needs gofmt:"; echo "$unformatted"; exit 1; fi
    go vet ./...
    go test -race ./...

# Set the version in flake.nix before tagging a release (just bump 0.2.0)
bump new_version:
    #!/usr/bin/env bash
    set -euo pipefail
    v="{{ new_version }}"
    v="${v#v}"
    sed -i -E "s/^( *version = )\"[^\"]*\";/\1\"$v\";/" flake.nix
    grep -qF "version = \"$v\";" flake.nix || { echo "could not set version in flake.nix" >&2; exit 1; }
    echo "flake.nix version is now $v; commit, then: git tag -a v$v -m v$v"

clean:
    rm -f tasked
