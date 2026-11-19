#!/usr/bin/env bash
# Builds build/Cais.app, a universal (Apple silicon + Intel) macOS app.
# Run `make mac` rather than this directly: it builds the frontend first.
#
#   VERSION=1.2.0 BACKEND_URL=https://cais.example ./scripts/build-macos.sh
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-1.0.0}"
APP="build/Cais.app"

if [ ! -f web/dist/index.html ]; then
  echo "web/dist is empty: run 'make frontend' first" >&2
  exit 1
fi

rm -rf "$APP" build/Cais.iconset
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

ldflags="-s -w"
if [ -n "${BACKEND_URL:-}" ]; then
  ldflags="$ldflags -X main.defaultBackend=$BACKEND_URL"
fi

for arch in arm64 amd64; do
  CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" CGO_CXXFLAGS="-Wno-deprecated-literal-operator" \
    go build -ldflags "$ldflags" -o "build/cais-$arch" ./cmd/cais-mac
done
lipo -create -output "$APP/Contents/MacOS/Cais" build/cais-arm64 build/cais-amd64
rm build/cais-arm64 build/cais-amd64

go run ./tools/icons -iconset build/Cais.iconset
iconutil -c icns build/Cais.iconset -o "$APP/Contents/Resources/Cais.icns"
rm -rf build/Cais.iconset

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>Cais</string>
  <key>CFBundleDisplayName</key><string>Cais</string>
  <key>CFBundleIdentifier</key><string>io.github.heldermartins1.cais</string>
  <key>CFBundleExecutable</key><string>Cais</string>
  <key>CFBundleIconFile</key><string>Cais</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>${VERSION}</string>
  <key>CFBundleVersion</key><string>${VERSION}</string>
  <key>CFBundleDevelopmentRegion</key><string>pt-PT</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSApplicationCategoryType</key><string>public.app-category.travel</string>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSHumanReadableCopyright</key><string>App não oficial, sem ligação ao Metropolitano de Lisboa.</string>
  <key>NSAppTransportSecurity</key>
  <dict>
    <!-- The window loads the app's own server on 127.0.0.1 over plain http. -->
    <key>NSAllowsLocalNetworking</key><true/>
  </dict>
</dict>
</plist>
PLIST

# Ad-hoc signature: enough to run on this Mac. Sharing it with others needs a
# Developer ID signature and notarization.
codesign --force --deep --sign - "$APP"

echo "Built $APP ($VERSION)"
