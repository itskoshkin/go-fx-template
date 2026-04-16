#!/bin/bash

set -e

APP_NAME="go-fx-template"
BUILD_DIR="build"
CONFIG_FILE="config.yaml"
EXAMPLE_CONFIG="example_config.yaml"
PLATFORMS=(
  "linux/amd64"
  "windows/amd64"
  "darwin/amd64"
  "darwin/arm64"
)

trap 'echo "❌ Build failed at line $LINENO: \"$BASH_COMMAND\" (exit code: $?)" >&2; exit 1' ERR
trap 'if [ $? -eq 0 ]; then echo "✅ Finished, all builds are in $BUILD_DIR/ folder"; fi' EXIT

rm -rf "$BUILD_DIR"; mkdir -p "$BUILD_DIR" # Clean build dir
for PLATFORM in "${PLATFORMS[@]}"; do
  IFS="/" read -r GOOS GOARCH <<< "$PLATFORM" # Split PLATFORM into GOOS and GOARCH by "/"
  OUTPUT_DIR="$BUILD_DIR/${APP_NAME}_${GOOS}_${GOARCH}" # Set output directory for this platform
  mkdir -p "$OUTPUT_DIR" # Create output directory if it doesn't exist

  BINARY_NAME="${APP_NAME}-${GOARCH}" # Set name for binary
  if [ "$GOOS" == "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe" # Add .exe extension for Windows (must die)
  fi

   # Build Go binary for target platform
  echo "🔨 Building for $GOOS/$GOARCH..."
  env GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "$OUTPUT_DIR/$BINARY_NAME" ./cmd/main.go

  # Copy config
  if [ -f "$CONFIG_FILE" ]; then
    cp "$CONFIG_FILE" "$OUTPUT_DIR/config.yaml"
  elif [ -f "$EXAMPLE_CONFIG" ]; then
    cp "$EXAMPLE_CONFIG" "$OUTPUT_DIR/config.yaml"
    echo "⚠️ File $CONFIG_FILE not found, used $EXAMPLE_CONFIG instead"
  else
    echo "⚠️ No config file found, skipping"
  fi

  # Pack to zip
  ZIP_NAME="$BUILD_DIR/${APP_NAME}_${GOOS}_${GOARCH}.zip"
  (cd "$OUTPUT_DIR" && zip -r "../$(basename "$ZIP_NAME")" . > /dev/null) # Create zip archive from output directory contents
  echo "📦 Packed $ZIP_NAME"
done
