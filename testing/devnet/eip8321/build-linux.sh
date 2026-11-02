#!/usr/bin/env bash
set -euo pipefail
export GOPATH=/cache/gopath GOCACHE=/cache/go-build GOMODCACHE=/cache/gopath/pkg/mod
if [[ ! -x /cache/bazel ]]; then
  curl --fail --location --retry 3 https://github.com/bazelbuild/bazel/releases/download/7.4.1/bazel-7.4.1-linux-x86_64 -o /cache/bazel
  chmod +x /cache/bazel
fi
/cache/bazel --batch --output_user_root=/cache/bazel-root build \
  //cmd/beacon-chain:beacon-chain //cmd/validator:validator --config=minimal \
  --workspace_status_command= --symlink_prefix=/cache/bazel-links/ --jobs=4 --enable_runfiles=false
mkdir -p /out/bin
cp --remove-destination /cache/bazel-links/bin/cmd/beacon-chain/beacon-chain_/beacon-chain /out/bin/beacon-chain
cp --remove-destination /cache/bazel-links/bin/cmd/validator/validator_/validator /out/bin/validator
sha256sum /out/bin/beacon-chain /out/bin/validator > /out/binaries.sha256
