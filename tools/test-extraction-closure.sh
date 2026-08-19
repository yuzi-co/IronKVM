#!/bin/sh
# Check that every test which lifts shell out of a device script lifts enough
# of it to run.
#
#   test-extraction-closure.sh [repo-root]
#
# Most suites here test the shipped text rather than a copy of it: they cut a
# function or a marked block out of a script under kvmapp/ or tools/, write it
# to a file, source that file and call into it. That is the right shape. It has
# one failure mode, and the failure mode is silent.
#
# If the lifted text calls a helper that was left behind, the shell reports
# "not found" and returns 127. Nothing stops, and 127 is falsy, so:
#
#   a case asserting the feature happens     FAILS, on a script that is correct
#   a case asserting it does NOT happen      PASSES, having tested nothing
#
# The second is the dangerous one, because it reads green for as long as it
# exists.
#
# This is not hypothetical. tools/usbdev/test-acm-console.sh lifted
# start_usb_dev on its own. That function calls usb_has, usb_resolve,
# usb_dropped, usb_prune_list and usb_report, and those call five more again.
# Two cases failed and three passed vacuously, while the board had acm.GS0
# linked into its gadget config and /dev/ttyGS0 present the whole time.
#
# What this does: for each suite, gather every extraction it writes to a file,
# work out which functions are then callable, and fail if the lifted text calls
# a function that its source script defines and the harness does not have.
#
# Three things stop it reporting a fault that cannot happen. All of a suite's
# extractions are judged together, because a suite commonly lifts two or three
# marked blocks and sources them as one file. Functions the suite defines
# itself count as available, because a deliberate stub is a definition like any
# other. Comment lines are dropped before anything counts as a call, since a
# block that explains which guard disarms it names that guard in prose.
#
# Two things stop it going quiet, which matters more. A suite whose extraction
# cannot be parsed is reported, not skipped: an earlier version of this file
# skipped tools/usbdev/test-acm-console.sh because its extraction spans a line
# continuation, then announced that everything was closed. And the count of
# examined suites has a floor, because a check that recognises nothing passes
# forever, which is the fault this file exists to catch.
#
# A suite that takes its source script only as a command line argument is
# listed as SKIP. Nothing static can tell what it will be handed.
set -u

ROOT=${1:-$(cd "$(dirname "$0")/.." && pwd)}
[ -d "$ROOT/tools" ] || { echo "usage: test-extraction-closure.sh [repo-root]"; exit 1; }

fails=0
checked=0
skipped=0
note() { printf '  %-70s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# The shell functions a file declares. Both "name() {" and "name(){" appear in
# this tree, and matching only the first is the mistake that hid the original
# fault once already.
defined_names() {
    grep -oE '^[a-z_][a-z_0-9]*\(\) *\{' "$1" 2>/dev/null | sed 's/().*//' | sort -u
}

# A shell line continued with a trailing backslash is one command. Join them
# before looking for anything, or an extraction whose redirection sits two
# lines below its sed program is invisible. awk rather than sed, because the
# backslash handling stays readable.
join_continuations() {
    awk '{ line = $0
           while (sub(/\\$/, "", line)) {
               if ((getline nxt) > 0) { line = line " " nxt } else { break }
           }
           print line }' "$1"
}

echo "===== every lifted block carries the helpers it calls ====="

: > "$WORK/verdicts"

for suite in $(find "$ROOT/tools" -name 'test-*.sh' | sort); do
    rel=${suite#"$ROOT"/}
    sdir=$(cd "$(dirname "$suite")" && pwd)
    : > "$WORK/lifted.sh"
    : > "$WORK/sources.txt"
    : > "$WORK/unparsed"
    : > "$WORK/skips"

    join_continuations "$suite" > "$WORK/joined.sh"

    # Only extractions redirected into a file are examined. Those get sourced
    # and run. A $(sed ... | grep ...) pipeline is a text assertion, where a
    # missing helper cannot return 127 because nothing executes. The redirect
    # need not follow the source directly: a pipe through another sed is
    # common, so anything without its own redirect may sit between.
    grep -E "sed -n '[^']*'[^\"]*\"\\\$[A-Za-z_][A-Za-z_0-9]*\"[^>]*>" "$WORK/joined.sh" 2>/dev/null \
    | while IFS= read -r line; do
        # A sed inside a command substitution is a text assertion: its output
        # becomes a value to compare, and nothing executes it. Those lines grep
        # the lifted text for things like `: > "$BOOT/recovery"`, which puts a
        # greater-than sign inside a quoted pattern and made an earlier version
        # of this check read it as a redirect.
        case "$line" in
            *'$(sed -n'*) continue ;;
        esac

        # The last "$VAR" that still sits before a redirect is the source. A
        # greedy match is deliberate: a sed program may itself contain a double
        # quote, as one suite's does, so anchoring on the first quote finds the
        # middle of the program instead of the file it reads.
        var=$(printf '%s\n' "$line" | sed -n 's/.*"\$\([A-Za-z_][A-Za-z_0-9]*\)"[^>]*>.*/\1/p')
        if [ -z "$var" ]; then
            echo "no source variable on an extraction line" >> "$WORK/unparsed"
            continue
        fi

        # Resolve the variable to a path. Two forms appear here:
        #   VAR=${1:-$(dirname "$0")/../../kvmapp/...}
        #   VAR="$HERE/../../kvmapp/..."      with HERE=$(cd "$(dirname "$0")" && pwd)
        value=$(grep -E "^$var=" "$suite" | head -1 | sed "s/^$var=//")
        value=${value%\}}
        value=${value#\$\{1:-}
        value=$(printf '%s\n' "$value" | sed 's/^"//; s/"$//')
        value=$(printf '%s\n' "$value" | sed 's|\$(dirname "\$0")|.|g; s|\${HERE}|.|g; s|\$HERE|.|g')

        case "$value" in
            *'$'* | '')
                echo "\$$var is only known at run time" >> "$WORK/skips"
                continue ;;
        esac

        prog=$(printf '%s\n' "$line" | sed -n "s/.*sed -n '\\([^']*\\)'.*/\\1/p")
        if [ -z "$prog" ]; then
            echo "no sed program on an extraction line naming \$$var" >> "$WORK/unparsed"
            continue
        fi

        dir=$(cd "$sdir" 2>/dev/null && cd "$(dirname "$value")" 2>/dev/null && pwd)
        src="$dir/$(basename "$value")"
        if [ -z "$dir" ] || [ ! -f "$src" ]; then
            echo "\$$var points at $value, which is not a file" >> "$WORK/unparsed"
            continue
        fi

        # Where the extraction is written, and whether the suite then sources
        # or runs that file. This is the whole test for "does it execute".
        # A greater-than sign alone is not enough: one suite greps the lifted
        # text for a literal redirect, so the character appears inside a quoted
        # pattern and means nothing.
        target=$(printf '%s\n' "$line" | sed -n 's|.*>>* *"\([^"]*\)".*|\1|p')
        [ -n "$target" ] || continue

        # The file has to be named a second time somewhere in the suite. A
        # lifted file that is written and never mentioned again is not run.
        #
        # Counting mentions rather than matching a sourcing form on purpose.
        # Suites source these as `. "$W/f.sh"`, as `sh "$W/f.sh"`, and as
        # `sh -c ". $W/f.sh; ..."` with the quotes escaped inside another
        # string. Enumerating those forms is how the previous attempt quietly
        # stopped examining two suites that were fine.
        [ "$(grep -cF "$target" "$suite")" -ge 2 ] || continue

        sed -n "$prog" "$src" >> "$WORK/lifted.sh" 2>/dev/null
        echo "$src" >> "$WORK/sources.txt"
    done

    if [ -s "$WORK/unparsed" ]; then
        echo "FAIL $rel: $(head -1 "$WORK/unparsed")" >> "$WORK/verdicts"
        continue
    fi

    if [ ! -s "$WORK/sources.txt" ]; then
        [ -s "$WORK/skips" ] && echo "SKIP $rel: $(head -1 "$WORK/skips")" >> "$WORK/verdicts"
        continue
    fi

    if [ ! -s "$WORK/lifted.sh" ]; then
        echo "FAIL $rel: an extraction produced nothing, so it tests an empty file" >> "$WORK/verdicts"
        continue
    fi

    grep -v '^[[:space:]]*#' "$WORK/lifted.sh" > "$WORK/code.sh"

    # A suite can define a helper inside a heredoc or an echo, so match a
    # definition anywhere in its text rather than only at the start of a line.
    { defined_names "$WORK/code.sh"
      grep -oE '[a-z_][a-z_0-9]*\(\) *\{' "$suite" 2>/dev/null | sed 's/().*//'
    } | sort -u > "$WORK/have.txt"

    missing=
    for src in $(sort -u "$WORK/sources.txt"); do
        for name in $(defined_names "$src"); do
            grep -qx "$name" "$WORK/have.txt" && continue
            grep -qE "(^|[^a-zA-Z0-9_])$name([^a-zA-Z0-9_(]|\$)" "$WORK/code.sh" || continue
            case " $missing " in
                *" $name "*) ;;
                *) missing="$missing $name" ;;
            esac
        done
    done

    from=$(sort -u "$WORK/sources.txt" | while read -r s; do printf '%s ' "$(basename "$s")"; done)
    if [ -n "$missing" ]; then
        echo "FAIL $rel lifts from $from without:$missing" >> "$WORK/verdicts"
    else
        echo "OK $rel <- $from" >> "$WORK/verdicts"
    fi
done

while IFS= read -r v; do
    [ -n "$v" ] || continue
    case "$v" in
        "OK "*)   checked=$((checked + 1)); note "${v#OK }" OK ;;
        "FAIL "*) checked=$((checked + 1)); note "${v#FAIL }" FAIL ;;
        "SKIP "*) skipped=$((skipped + 1)); printf '  %-70s %s\n' "${v#SKIP }" SKIP ;;
    esac
done < "$WORK/verdicts"

# A floor, so this cannot pass by recognising nothing. It sits well under the
# current count: it catches the detector breaking outright, and it is not a
# number to keep raising as suites are added.
if [ "$checked" -lt 12 ]; then
    note "at least twelve harnesses were examined (found $checked)" FAIL
    echo "  (matching almost none means this check stopped recognising the suites)"
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all $checked harness(es) are closed, $skipped skipped"
else
    echo "$fails of $checked harness(es) FAILED, $skipped skipped"
    exit 1
fi
