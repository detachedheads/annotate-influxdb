#!/usr/bin/env bash
set -e

# Get the base repository path
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
# Get the name of the package we are building
PACKAGE="$(basename "$DIR")"
# Get the name of the organization
ORGANIZATION="$(basename "$(dirname "$DIR")")"
# Get the name of the repository
REPOSITORY="$(basename "$(dirname "$(dirname "$DIR")")")"

: ${LOCAL_TARGET="$(go env GOOS)_$(go env GOARCH)"}

# Move into our base repository path
cd "$DIR"

# Get the version of the app
VERSION="$(cat VERSION)"

# Clean up old binaries and packages
echo "==> Cleaning up build environment..."
rm -rf pkg/*
rm -rf bin/*
mkdir -p bin
mkdir -p pkg

#
# Compile Configuration
#

GIT_COMMIT="$(git rev-parse --short HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"
EXTLDFLAGS="-X $REPOSITORY/$ORGANIZATION/$PACKAGE/cmd.GITCOMMIT=${GIT_COMMIT}${GIT_DIRTY} -X $REPOSITORY/$ORGANIZATION/$PACKAGE/cmd.VERSION=$VERSION"
STATIC="-extldflags '-static'"

#
# Determine build targets
#

# Default to local os/arch
targets="$LOCAL_TARGET"

# If we are building for release change targets based off of environment
if [[ "$TARGETS" == "release" ]]; then
  if [[ $(uname) == "Linux" ]]; then
    targets="darwin_amd64 linux_amd64 linux_amd64-lxc windows_386 windows_amd64"
  elif [[ $(uname) == "Darwin" ]]; then
    targets="darwin_amd64 linux_amd64 linux_amd64-lxc"
  else
    echo "Unable to build on $(uname). Use Linux or Darwin."
    exit 1
  fi
elif [[ "$TARGETS" != "" ]]; then
  targets="$TARGETS"
fi

set +e

for target in $targets; do
  case $target in
    "darwin_amd64")
      echo "==> Building darwin amd64..."
      CGO_ENABLED=0 GOARCH="amd64" GOOS="darwin" \
        go build -ldflags "$EXTLDFLAGS" -o "pkg/darwin_amd64/$PACKAGE"
      ;;
    "linux_amd64")
      echo "==> Building linux amd64..."
      CGO_ENABLED=1 GOOS="linux" GOARCH="amd64" \
        go build -ldflags "$STATIC $EXTLDFLAGS" -o "pkg/linux_amd64/$PACKAGE"
      ;;
    "linux_amd64-lxc")
      echo "==> Building linux amd64 with lxc..."
      CGO_ENABLED=1 GOOS="linux" GOARCH="amd64" \
        go build -ldflags "$STATIC $EXTLDFLAGS" -o "pkg/linux_amd64-lxc/$PACKAGE" -tags "lxc"
      ;;
    "windows_386")
      echo "==> Building windows 386..."
      CGO_ENABLED=1 GOOS="windows" GOARCH="386" CXX="i686-w64-mingw32-g++" CC="i686-w64-mingw32-gcc" \
        go build -ldflags "$STATIC $EXTLDFLAGS" -o "pkg/windows_386/$PACKAGE.exe"
      ;;
    "windows_amd64")
      echo "==> Building windows amd64..."
      CGO_ENABLED=1 GOOS="windows"  GOARCH="amd64" CXX="x86_64-w64-mingw32-g++" CC="x86_64-w64-mingw32-gcc" \
        go build -ldflags "$STATIC $EXTLDFLAGS" -o "pkg/windows_amd64/$PACKAGE.exe"
      ;;
    *)
      echo "--> Invalid target: $target"
      ;;
  esac
done

set -e

# Copy our local OS/Arch to the bin/ directory
for F in $(find ./pkg/${LOCAL_TARGET} -mindepth 1 -maxdepth 1 -type f); do
  echo "==> Copying ${LOCAL_TARGET} to ./bin"
  cp ${F} bin/
  chmod 755 bin/*
done

# Package up the artifacts
if [[ "$PACKAGE" != "" ]]; then
  for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})
    echo "==> Packaging ${OSARCH}"
    pushd $PLATFORM >/dev/null 2>&1
    tar czvf ../${PACKAGE}-${OSARCH}.tar.gz ./* >/dev/null
    popd >/dev/null 2>&1
  done

  echo "==> Generating SHA256..."
  for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type f); do
    OSARCH=$(basename ${PLATFORM})
    # Need to make this portable
    shasum -a 256 "./pkg/${PACKAGE}-${OSARCH}" >> ./pkg/SHASUM256.txt
  done

  cat ./pkg/SHASUM256.txt
fi