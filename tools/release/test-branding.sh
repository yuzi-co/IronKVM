#!/bin/sh
# The branding surface, guarded so a later edit cannot quietly undo it.
#
# There is no frontend test runner in this repository, so these are grep checks.
# They are worth having anyway: the interesting half is not that the new name
# appears, it is that the hardware references did NOT get caught in a blanket
# search and replace. The board is a NanoKVM whatever firmware it runs.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../.." && pwd)
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

echo "branding"

check "the page title is IronKVM" \
    "$(grep -c '<title>IronKVM</title>' "$ROOT/web/index.html")" "1"
check "the favicon is the IronKVM mark" \
    "$(grep -c 'ironkvm.ico' "$ROOT/web/index.html")" "1"
check "the login page uses the IronKVM mark" \
    "$(grep -c 'ironkvm.ico' "$ROOT/web/src/pages/auth/login/index.tsx")" "1"
check "the vendor icon is gone" \
    "$([ -e "$ROOT/web/public/sipeed.ico" ] && echo present || echo gone)" "gone"

# S95nanokvm replaces the served icon with /boot/logo.ico when an operator puts
# one there. Renaming the icon without renaming it here would leave that feature
# swapping a file nothing loads, and it would fail silently: the page would keep
# working and simply ignore the operator's logo.
check "the custom logo override follows the icon name" \
    "$(grep -c 'web/ironkvm.ico' "$ROOT/kvmapp/system/init.d/S95nanokvm")" "3"
check "the custom logo override no longer names the vendor icon" \
    "$(grep -c 'web/sipeed.ico' "$ROOT/kvmapp/system/init.d/S95nanokvm")" "0"
check "the default web title sentinel is IronKVM" \
    "$(grep -c '"IronKVM"' "$ROOT/server/service/vm/web_title.go")" "1"
check "the default update server is the IronKVM feed" \
    "$(grep -c 'yuzi-co.github.io/IronKVM' "$ROOT/server/service/application/service.go")" "1"
check "the Sipeed feed is kept so a board can leave" \
    "$(grep -c 'cdn.sipeed.com/nanokvm\"' "$ROOT/server/service/application/service.go")" "1"
check "the update page knows the Sipeed feed" \
    "$(grep -c 'const SIPEED_UPDATE_SERVER' "$ROOT/web/src/pages/desktop/menu/settings/update/custom-server.tsx")" "1"
check "the update page offers it as a one-click preset" \
    "$(grep -c 'customServer.useSipeed' "$ROOT/web/src/pages/desktop/menu/settings/update/custom-server.tsx")" "1"
check "the preset has an English label" \
    "$(grep -c 'useSipeed:' "$ROOT/web/src/i18n/locales/en.ts")" "1"
check "the About panel carries the disclaimer" \
    "$(grep -c 'Not affiliated with Sipeed' "$ROOT/web/src/pages/desktop/menu/settings/about/community.tsx")" "1"
check "the About panel still links the hardware vendor" \
    "$(grep -c 'wiki.sipeed.com' "$ROOT/web/src/pages/desktop/menu/settings/about/community.tsx")" "2"

# Hardware references must survive. A blanket search and replace across the
# locale files is the failure this guards against.
check "the reset instruction still names the hardware" \
    "$(grep -c 'BOOT button on the NanoKVM' "$ROOT/web/src/i18n/locales/en.ts")" "2"
# The wording changed when the card image gained a version of its own: the field
# now shows both, so the tooltip has to name both. What this guards is unchanged
# - the hardware reference must survive, because the system image is Sipeed's
# and calling it anything else would be a lie about whose kernel is running.
check "the system image version still names NanoKVM" \
    "$(grep -c 'NanoKVM system image' "$ROOT/web/src/i18n/locales/en.ts")" "1"
check "the tooltip names the card image too" \
    "$(grep -c 'IronKVM card image' "$ROOT/web/src/i18n/locales/en.ts")" "1"
check "the product strings were changed" \
    "$(grep -c 'About IronKVM' "$ROOT/web/src/i18n/locales/en.ts")" "1"

# Only the English source is retranslated. A part-translated locale file is
# worse than an untouched one, because the reader cannot tell which half is
# current.
check "no locale but en.ts mentions IronKVM" \
    "$(grep -l 'IronKVM' "$ROOT"/web/src/i18n/locales/*.ts 2>/dev/null | grep -vc '/en\.ts$')" "0"

# The offline upload validates the file name in the browser, before the server
# ever sees it. The server pattern was widened to accept both products and this
# one was not, so the page refused every IronKVM release with "Invalid filename
# format. Please download from GitHub releases." and the link beside it pointed
# at Sipeed's releases, which do not carry one.
OFFLINE="$ROOT/web/src/pages/desktop/menu/settings/update/offline.tsx"
check "the offline upload accepts an IronKVM package" \
    "$(grep -c 'nanokvm|ironkvm' "$OFFLINE")" "1"
check "the offline upload still accepts an official package" \
    "$(grep -c 'nanokvm|ironkvm' "$OFFLINE")" "1"
check "the releases link points at this fork" \
    "$(grep -c 'github.com/yuzi-co/IronKVM/releases' "$OFFLINE")" "1"
check "the releases link no longer points at Sipeed" \
    "$(grep -c 'github.com/sipeed/NanoKVM/releases' "$OFFLINE")" "0"

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
