#!/bin/sh
# Installs the Ceryx CLI from GitHub releases:
#   curl -fsSL https://raw.githubusercontent.com/withlogicco/ceryx-cli/main/install.sh | sh
# Env: CERYX_VERSION (tag, default latest), INSTALL_DIR (default /usr/local/bin).
set -eu

repo=withlogicco/ceryx-cli
dir=${INSTALL_DIR:-/usr/local/bin}

os=$(uname -s | tr A-Z a-z)
case $os in darwin | linux) ;; *) echo "unsupported OS: $os" >&2; exit 1 ;; esac
arch=$(uname -m)
case $arch in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

if [ -n "${CERYX_VERSION:-}" ]; then
	url=https://github.com/$repo/releases/download/$CERYX_VERSION/ceryx_${os}_${arch}
else
	url=https://github.com/$repo/releases/latest/download/ceryx_${os}_${arch}
fi

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsSL -o "$tmp" "$url"
chmod 755 "$tmp"

sudo=
[ -w "$dir" ] || sudo=sudo
$sudo mkdir -p "$dir"
$sudo mv "$tmp" "$dir/ceryx"
echo "installed $("$dir/ceryx" --version) to $dir/ceryx"
