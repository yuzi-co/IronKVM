#!/bin/sh
# Prove that test-deploy-guard.sh fails when the probe URL is built wrongly.
#
#   test-deploy-guard-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped script is only ever read.
#
# This one is worth breaking on purpose because of what it decides. The guard
# restores the previous binary when the probe does not answer 200, so a probe
# aimed at the wrong scheme or the wrong port does not report a problem, it
# creates one: it rolls a working server back and tells the operator the deploy
# failed. Cases that only assert "a URL came out" pass on every one of those.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the cases pass because there is nothing wrong with it, and the
# run reads as proof when it proved nothing.
DIR=$(dirname "$0")
DG=${1:-$DIR/deploy-server}
TEST=$DIR/test-deploy-guard.sh
for f in "$DG" "$TEST"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 1; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$DG" "$d/deploy-server"
    "$@" "$d/deploy-server"

    if cmp -s "$d/deploy-server" "$DG"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if timeout 60 sh "$TEST" "$d/deploy-server" > /dev/null 2>&1; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The defect as it shipped: the probe is always plain HTTP. On a board serving
# HTTPS that answers 307, and the guard restores the old binary over a good one.
m_fixedhttp() {
    sed -i 's@^DEPLOY_URL=${DEPLOY_URL:-$(default_probe_url)}$@DEPLOY_URL=${DEPLOY_URL:-http://127.0.0.1/}@' "$1"
}

# The scheme is read but the port is not, so a server on a port of its own is
# probed on the default one and reads as dead.
m_noport() {
    sed -i 's@^        \*)                echo "$proto://127.0.0.1:$port/" ;;$@        *)                echo "$proto://127.0.0.1/" ;;@' "$1"
}

# The port lookup stops caring which scheme is in use and takes http every time.
m_wrongscheme() {
    sed -i 's@\$proto:\[\[:space:\]\]\*@http:[[:space:]]*@' "$1"
}

# The lookup is no longer scoped to the port block, so the first key of that
# name anywhere in the file wins.
m_unscoped() {
    sed -i "s@sed -n '/^port:/,/^\[^\[:space:\]#\]/p' \"\$conf\" \\\\@cat \"\$conf\" | sed -n 'p' \\\\@" "$1"
}

# DEPLOY_URL stops being overridable, so an operator cannot point the probe at
# a board that answers somewhere else.
m_notoverridable() {
    sed -i 's@^DEPLOY_URL=${DEPLOY_URL:-$(default_probe_url)}$@DEPLOY_URL=$(default_probe_url)@' "$1"
}

# A config that cannot be read stops falling back, so a board with no config
# file is probed at an empty URL.
m_nofallback() {
    sed -i 's@^    proto=http$@    proto=@' "$1"
}

echo "===== every mutation must be caught ====="
try "the probe goes back to fixed http"        m_fixedhttp
try "a non-default port is dropped"            m_noport
try "the port is read for the wrong scheme"    m_wrongscheme
try "the port lookup leaves the port block"    m_unscoped
try "DEPLOY_URL stops being overridable"       m_notoverridable
try "an unreadable config has no fallback"     m_nofallback

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
