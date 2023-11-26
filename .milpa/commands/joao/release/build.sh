#!/usr/bin/env bash
# Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
# SPDX-License-Identifier: Apache-2.0

platforms=(
  linux/amd64
  linux/arm64
  linux/arm
  linux/mips
  linux/mips64
  darwin/arm64
  darwin/amd64
)
root=$(dirname "$MILPA_COMMAND_REPO")
cd "$root" || @milpa.fail "could not cd into $root"

@milpa.log info "Starting build for version $MILPA_ARG_VERSION"

for platform in "${platforms[@]}"; do
  @milpa.log info "building for $platform"
  os="${platform%%/*}"
  arch="${platform##*/}"
  base="dist/$os-$arch"
  mkdir -p "$base" || @milpa.fail "Could not create dist dir"
  GOOS="$os" GOARCH="$arch" go build -ldflags "-s -w -X git.rob.mx/nidito/joao/pkg/version.Version=$MILPA_ARG_VERSION" -trimpath -o "$base/joao"
  @milpa.log success "built for $platform"

  package="$root/dist/joao-$os-$arch.tgz"
  @milpa.log info "archiving to $package"
  (cd "$base" && tar -czf "$package" joao) || @milpa.fail "Could not archive $package"
  openssl dgst -sha256 "$package" | awk '{print $2}' > "$package.shasum"
  rm -rf "$base"
  @milpa.log success "archived $package"
done

echo -n "$MILPA_ARG_VERSION" > "$root/dist/latest-version"

@milpa.log info "uploading to cdn"
rclone sync --s3-acl=public-read \
  "$root/dist/" \
  "cdn:cdn.rob.mx/tools/joao/$MILPA_ARG_VERSION/" || @milpa.fail "could not upload to CDN"
@milpa.log complete "release for $MILPA_ARG_VERSION available at CDN"
rm -rf dist
