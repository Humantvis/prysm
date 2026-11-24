#!/usr/bin/env bash
set -euo pipefail
cache=${CACHE_DIR:-/cache}
out=${OUTPUT_DIR:-/out}
export GOPATH="$cache/gopath" GOCACHE="$cache/go-build" GOMODCACHE="$cache/gopath/pkg/mod"
case "$(uname -m)" in
  x86_64) arch=x86_64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo 'Unsupported build architecture' >&2; exit 1 ;;
esac
bazel="$cache/bazel-7.4.1-$arch"
mkdir -p "$cache" "$out/bin"
if [[ ! -x "$bazel" ]]; then
  url="https://github.com/bazelbuild/bazel/releases/download/7.4.1/bazel-7.4.1-linux-$arch"
  curl --fail --location --retry 3 "$url" -o "$bazel"
  curl --fail --location --retry 3 "$url.sha256" -o "$bazel.sha256"
  echo "$(cut -d ' ' -f 1 "$bazel.sha256")  $bazel" | sha256sum --check
  chmod +x "$bazel"
fi
"$bazel" --batch --output_user_root="$cache/bazel-root" build \
  //cmd/beacon-chain:beacon-chain //cmd/validator:validator --config=minimal \
  --workspace_status_command= --symlink_prefix="$cache/bazel-links/" --jobs="${BUILD_JOBS:-4}" --enable_runfiles=false
cp --remove-destination "$cache/bazel-links/bin/cmd/beacon-chain/beacon-chain_/beacon-chain" "$out/bin/beacon-chain"
cp --remove-destination "$cache/bazel-links/bin/cmd/validator/validator_/validator" "$out/bin/validator"
sha256sum "$out/bin/beacon-chain" "$out/bin/validator" > "$out/binaries.sha256"
