#!/bin/sh
set -eu

repository="${GPC_REPO:-aljrico/Google-Play-Console-CLI}"
version="${GPC_VERSION:-latest}"
install_dir="${GPC_INSTALL_DIR:-/usr/local/bin}"
binary_name="gpc"

usage() {
	cat <<'USAGE'
Install gpc from GitHub Releases.

Environment:
  GPC_VERSION       Release tag to install, for example v0.1.0. Defaults to latest.
  GPC_INSTALL_DIR  Directory for the gpc binary. Defaults to /usr/local/bin.
  GPC_REPO         GitHub repository. Defaults to aljrico/Google-Play-Console-CLI.
  GPC_BASE_URL     Override release asset URL base, mainly for tests.
USAGE
}

fail() {
	printf 'gpc install: %s\n' "$1" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

download() {
	url="$1"
	output="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$output"
		return
	fi
	if command -v wget >/dev/null 2>&1; then
		wget -q "$url" -O "$output"
		return
	fi
	fail "curl or wget is required"
}

github_api() {
	path="$1"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "https://api.github.com/repos/$repository/$path"
		return
	fi
	if command -v wget >/dev/null 2>&1; then
		wget -q "https://api.github.com/repos/$repository/$path" -O -
		return
	fi
	fail "curl or wget is required"
}

detect_os() {
	case "$(uname -s)" in
		Darwin) printf 'Darwin' ;;
		Linux) printf 'Linux' ;;
		*) fail "unsupported OS: $(uname -s)" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64 | amd64) printf 'x86_64' ;;
		arm64 | aarch64) printf 'arm64' ;;
		*) fail "unsupported architecture: $(uname -m)" ;;
	esac
}

latest_tag() {
	github_api "releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1
}

sha256_file() {
	file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
		return
	fi
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
		return
	fi
	fail "sha256sum or shasum is required"
}

install_binary() {
	source_path="$1"
	target_path="$2"
	if [ ! -d "$install_dir" ]; then
		parent_dir="$(dirname "$install_dir")"
		if [ -w "$parent_dir" ]; then
			install -d "$install_dir"
		fi
	fi
	if [ -w "$install_dir" ]; then
		install -m 0755 "$source_path" "$target_path"
		return
	fi
	if command -v sudo >/dev/null 2>&1; then
		sudo install -d "$install_dir"
		sudo install -m 0755 "$source_path" "$target_path"
		return
	fi
	fail "$install_dir is not writable and sudo is unavailable"
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
	usage
	exit 0
fi

need awk
need dirname
need sed
need tar
need install

os="$(detect_os)"
arch="$(detect_arch)"

if [ "$version" = "latest" ]; then
	version="$(latest_tag)"
	[ -n "$version" ] || fail "could not resolve latest release tag"
fi

archive="gpc_${version#v}_${os}_${arch}.tar.gz"
base_url="${GPC_BASE_URL:-https://github.com/$repository/releases/download/$version}"

work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT INT TERM

archive_path="$work_dir/$archive"
checksums_path="$work_dir/checksums.txt"

download "$base_url/$archive" "$archive_path"
download "$base_url/checksums.txt" "$checksums_path"

expected_checksum="$(awk -v file="$archive" '$2 == file { print $1 }' "$checksums_path")"
[ -n "$expected_checksum" ] || fail "checksum for $archive not found"

actual_checksum="$(sha256_file "$archive_path")"
[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum mismatch for $archive"

tar -xzf "$archive_path" -C "$work_dir" "$binary_name"
install_binary "$work_dir/$binary_name" "$install_dir/$binary_name"

printf 'gpc installed to %s\n' "$install_dir/$binary_name"
