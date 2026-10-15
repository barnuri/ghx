#!/bin/sh
set -eu

repo="barnuri/ghx"
install_dir="${GHX_INSTALL_DIR:-$HOME/.local/bin}"
shim="${GHX_SHIM:-1}"
base_url="https://github.com/$repo/releases/latest/download"

fail() { echo "ghx install: $*" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) fail "unsupported OS $(uname -s). On Windows run install.ps1." ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) fail "unsupported CPU $(uname -m)" ;;
esac

download() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$2" "$1"
  else
    fail "curl or wget is required"
  fi
}

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

asset="ghx_${os}_${arch}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $asset..."
download "$base_url/$asset" "$tmp/ghx"
download "$base_url/checksums.txt" "$tmp/checksums.txt"
expected="$(awk -v name="$asset" '$2 == name { print $1 }' "$tmp/checksums.txt")"
[ -n "$expected" ] || fail "no checksum for $asset"
[ "$(sha256 "$tmp/ghx")" = "$expected" ] || fail "checksum mismatch for $asset"

mkdir -p "$install_dir"
chmod 755 "$tmp/ghx"
mv "$tmp/ghx" "$install_dir/ghx"
echo "Installed $install_dir/ghx"

if [ "$shim" = 1 ]; then
  if [ -e "$install_dir/gh" ] && [ ! -L "$install_dir/gh" ]; then
    echo "Skipped gh shim: $install_dir/gh is a real file. Use ghx, or move that file." >&2
  else
    ln -sf ghx "$install_dir/gh"
    echo "Linked $install_dir/gh -> ghx"
  fi
fi

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add to your shell profile: export PATH=\"$install_dir:\$PATH\"" ;;
esac
if [ "$shim" = 1 ] && [ "$(command -v gh 2>/dev/null)" != "$install_dir/gh" ]; then
  echo "Put $install_dir before other gh locations in PATH so 'gh' runs ghx."
fi
echo "Next: create ~/.config/ghx/config.yaml (see https://github.com/$repo#configuration), then run: gh ghx which"
