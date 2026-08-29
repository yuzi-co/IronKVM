#!/bin/sh
# Hold every GitHub action to a commit, not to a tag.
#
#   test-pinned-actions.sh [path-to-.github/workflows]
#
# Not destructive: it reads the workflow files, runs no build, and makes no
# network call.
#
# Why this test exists. `actions/checkout@v7` names a tag, and the account that
# owns the action can move that tag to another commit at any time. These
# workflows run with a checkout token, a GHCR login, and the release job's write
# permission in scope. A moved tag therefore runs somebody else's new code next
# to all three. tools/build/test-pinned-inputs.sh already holds the two builder
# images to verified downloads. Until 2026-08-29 the workflows were the last
# build input that a third party could still change under us.
#
# The pins do not follow new releases, and that is the point. Nothing here asks
# GitHub what the tag points at today. A test that compared the two would fail
# on the day an action published a release, and it would report a correct pin as
# a defect. Moving a pin is a deliberate edit. The `# vX.Y.Z` comment beside
# each sha records which version is in place, for a reader and for Dependabot.
HERE=$(cd "$(dirname "$0")" && pwd)
DIR=${1:-$HERE/../../.github/workflows}

[ -d "$DIR" ] || { echo "no workflow directory at $DIR"; exit 2; }
ROOT=$(cd "$DIR/../.." && pwd) || exit 2

set -- "$DIR"/*.yml
[ -f "$1" ] || { echo "no workflow file in $DIR"; exit 2; }

TMP=$(mktemp) || exit 2
REFS=$(mktemp) || exit 2
trap 'rm -f "$TMP" "$REFS"' EXIT INT TERM

TAB=$(printf '\t')
fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() {
    if [ "$2" = "$3" ]; then note "$1" OK; else note "$1 (got '$2', want '$3')" FAIL; fi
}

# One "<file> <line> <ref>" record per `uses:` step. Each file is read by name,
# which keeps the filename out of the parse. A path that holds a colon would
# otherwise split in the wrong place.
for f in "$DIR"/*.yml
do
    name=$(basename "$f")
    grep -n '^[[:space:]]*uses:' "$f" > "$TMP"
    while IFS= read -r line
    do
        [ -n "$line" ] || continue
        lineno=${line%%:*}
        ref=${line#*uses:}
        ref=$(printf '%s' "$ref" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
        printf '%s\t%s\t%s\n' "$name" "$lineno" "$ref" >> "$REFS"
    done < "$TMP"
done

total=$(wc -l < "$REFS" | tr -d ' ')
[ "$total" -gt 0 ] || { echo "no 'uses:' step under $DIR"; exit 2; }

echo "===== every action is pinned to a commit ====="
external=0
unpinned=0
offenders=
while IFS="$TAB" read -r name lineno ref
do
    case "$ref" in ./*|docker://*) continue ;; esac
    external=$((external + 1))
    printf '%s\n' "$ref" | grep -qE '@[0-9a-f]{40}([[:space:]]|$)' && continue
    unpinned=$((unpinned + 1))
    offenders="$offenders $name:$lineno"
done < "$REFS"
check "$external external reference(s), each pinned to a 40-character sha" "$unpinned" "0"
[ -n "$offenders" ] && echo "    not pinned:$offenders"

echo
echo "===== every pin says which version it is ====="
# Dependabot reads this comment to name the version it is moving away from, and
# a reader has no other way to tell 3d3c42e5 from any other commit.
nocomment=0
offenders=
while IFS="$TAB" read -r name lineno ref
do
    case "$ref" in ./*|docker://*) continue ;; esac
    printf '%s\n' "$ref" | grep -qE '@[0-9a-f]{40} # v[0-9]+\.[0-9]+\.[0-9]+$' && continue
    nocomment=$((nocomment + 1))
    offenders="$offenders $name:$lineno"
done < "$REFS"
check "every pin carries a '# vX.Y.Z' comment" "$nocomment" "0"
[ -n "$offenders" ] && echo "    no version comment:$offenders"

echo
echo "===== one action, one commit ====="
# actions/checkout is used six times. A bump that reaches five of them leaves
# the sixth on the old commit, both builds still pass, and nothing else here
# would report the difference.
split=$(awk -F'\t' '{print $3}' "$REFS" \
    | grep -vE '^(\./|docker://)' \
    | sed 's/[[:space:]]*#.*$//' \
    | sort -u \
    | awk -F'@' 'NF > 1 {print $1}' \
    | sort | uniq -d)
if [ -z "$split" ]
then
    note "no action is pinned to two different commits" OK
else
    for a in $split
    do
        note "$a is pinned to more than one commit" FAIL
    done
fi

echo
echo "===== a local reusable workflow resolves to a file ====="
locals=0
missing=0
while IFS="$TAB" read -r name lineno ref
do
    case "$ref" in ./*) ;; *) continue ;; esac
    locals=$((locals + 1))
    target=${ref%@*}
    [ -f "$ROOT/$target" ] && continue
    missing=$((missing + 1))
    note "$name:$lineno points at a missing $target" FAIL
done < "$REFS"
[ "$missing" -eq 0 ] && note "$locals local reference(s) resolve" OK

echo
if [ "$fails" -eq 0 ]
then
    echo "===== every action reference is pinned ====="
else
    echo "===== $fails check(s) failed ====="
    exit 1
fi
