#!/bin/sh
# Check that debug() prints the arguments it is given.
#
#   test-debug-passes-arguments.sh [kvm_vision.cpp]
#
# Not destructive: the source is only read, and the compiled part runs in a
# temporary directory.
#
# The fault this covers: debug() declared varargs and then called
# printf(format) with the format string alone. Of the calls in kvm_vision.cpp
# about a third carry a %d or a %s, and every one of those read arguments that
# were never passed, so the value printed was whatever the stack held. It was
# also a non-constant format string handed straight to printf, which is the
# shape -Wformat-security exists to find.
#
# Nothing had noticed because debug_en is 0 unless kvmv_init is asked for
# logging, and the server passes 0 always, so no call has ever printed on a
# device. That makes this a fault that only appears the first time somebody
# turns logging on to diagnose something else, which is the worst moment for a
# log to be wrong.
#
# Two things are checked, and they are separate. The definition has to consume
# the arguments, and the declaration has to carry the printf attribute so the
# compiler checks the call sites. Neither implies the other: a correct vprintf
# with no attribute still lets a wrong call site through silently.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

[ -f "$VIS" ] || { echo "missing: $VIS"; exit 2; }

CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-debug-passes-arguments.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# Anchored on the definition line itself. The shared body() helper used by the
# other suites keys on "name(", which an indented call site matches as well,
# and debug() is called a hundred times in this file.
debug_definition() {
    sed -n '/^void debug(const char \*format, \.\.\.)$/,/^}/p' "$1"
}

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

echo "===== the definition ====="

debug_body=$(debug_definition "$VIS")

printf '%s\n' "$debug_body" | grep -q 'va_start(args, format)' \
    && note "the definition starts a va_list" OK \
    || note "the definition starts a va_list" FAIL

printf '%s\n' "$debug_body" | grep -q 'vprintf(format, args)' \
    && note "the definition passes the arguments to vprintf" OK \
    || note "the definition passes the arguments to vprintf" FAIL

printf '%s\n' "$debug_body" | grep -q 'va_end(args)' \
    && note "the definition ends the va_list" OK \
    || note "the definition ends the va_list" FAIL

# The shape that was there. printf with the format alone is both the dropped
# argument and the format-security hole, so it is worth naming on its own
# rather than trusting the vprintf case above to exclude it.
printf '%s\n' "$debug_body" | grep -qE '[^v]printf\(format\)' \
    && note "the format string is not handed to printf alone" FAIL \
    || note "the format string is not handed to printf alone" OK

grep -q '#include <stdarg.h>' "$VIS" \
    && note "stdarg.h is included" OK \
    || note "stdarg.h is included" FAIL

echo "===== the declaration ====="

# Without this the compiler checks nothing, and a call whose arguments do not
# match its format string is as silent as the fault this replaces.
grep -q 'void debug(const char \*format, \.\.\.) __attribute__((format(printf, 1, 2)));' "$VIS" \
    && note "the declaration carries the printf attribute" OK \
    || note "the declaration carries the printf attribute" FAIL

echo "===== it behaves like printf ====="

# Lift the function out of the shipped file rather than restating it here, so
# a rewrite that changes the behaviour is caught by these cases and not only
# by the greps above.
#
# These five cases lock the behaviour in. They do not, on their own, detect the
# fault this suite was written for, and it is worth saying why rather than
# letting a later reader assume they do. Dropping the arguments is undefined
# behaviour, not deterministic corruption: on an ABI that passes the first few
# arguments in registers, printf reads the registers the caller happened to
# leave set and prints the right numbers anyway. Run against the version this
# replaces, all five pass. The static cases and the attribute case are what
# actually fail there.
{
    echo '#include <stdio.h>'
    echo '#include <stdarg.h>'
    echo '#include <string.h>'
    echo '#include <stdlib.h>'
    echo '#include <stdint.h>'
    echo 'uint8_t debug_en = 0;'
    printf '%s\n' "$debug_body"
    cat <<'MAIN'
int main(void)
{
    char buf[256];

    // Off by default, which is the state every device runs in.
    debug_en = 0;
    memset(buf, 0, sizeof(buf));
    setvbuf(stdout, buf, _IOFBF, sizeof(buf));
    debug("[t] silent %d\n", 41);
    fflush(stdout);
    setvbuf(stdout, NULL, _IONBF, 0);
    if (buf[0] != 0) {
        fprintf(stderr, "printed while disabled: %s\n", buf);
        return 1;
    }

    // On, one integer.
    debug_en = 1;
    memset(buf, 0, sizeof(buf));
    setvbuf(stdout, buf, _IOFBF, sizeof(buf));
    debug("[t] one %d\n", 41);
    fflush(stdout);
    setvbuf(stdout, NULL, _IONBF, 0);
    if (strcmp(buf, "[t] one 41\n") != 0) {
        fprintf(stderr, "one integer: got %s\n", buf);
        return 2;
    }

    // Several arguments of mixed type, which is what the busiest calls in
    // kvm_vision.cpp look like.
    memset(buf, 0, sizeof(buf));
    setvbuf(stdout, buf, _IOFBF, sizeof(buf));
    debug("[t] %s %d %d %x\n", "res", 1920, 1080, 255);
    fflush(stdout);
    setvbuf(stdout, NULL, _IONBF, 0);
    if (strcmp(buf, "[t] res 1920 1080 ff\n") != 0) {
        fprintf(stderr, "mixed: got %s\n", buf);
        return 3;
    }

    // No arguments at all, which is the other two thirds of the calls.
    memset(buf, 0, sizeof(buf));
    setvbuf(stdout, buf, _IOFBF, sizeof(buf));
    debug("[t] plain\n");
    fflush(stdout);
    setvbuf(stdout, NULL, _IONBF, 0);
    if (strcmp(buf, "[t] plain\n") != 0) {
        fprintf(stderr, "plain: got %s\n", buf);
        return 4;
    }

    // A percent sign that is data, not a specifier. The old code would have
    // read an argument for it.
    memset(buf, 0, sizeof(buf));
    setvbuf(stdout, buf, _IOFBF, sizeof(buf));
    debug("[t] %d%% busy\n", 90);
    fflush(stdout);
    setvbuf(stdout, NULL, _IONBF, 0);
    if (strcmp(buf, "[t] 90% busy\n") != 0) {
        fprintf(stderr, "percent: got %s\n", buf);
        return 5;
    }

    return 0;
}
MAIN
} > "$work/lifted.c"

if "$CXX" -x c++ -O1 -Wall -Wextra -Wformat=2 -o "$work/lifted" "$work/lifted.c" \
        2> "$work/build.log"; then
    note "the lifted function compiles" OK
    if "$work/lifted" > "$work/run.log" 2>&1; then
        note "it prints nothing while debug_en is 0" OK
        note "it prints one integer argument" OK
        note "it prints several arguments of mixed type" OK
        note "it prints a call that has no arguments" OK
        note "it prints a literal percent sign" OK
    else
        rc=$?
        note "the lifted function behaves like printf (exit $rc)" FAIL
        sed 's/^/    /' "$work/run.log"
    fi
else
    note "the lifted function compiles" FAIL
    sed 's/^/    /' "$work/build.log"
fi

echo "===== the compiler checks the call sites ====="

# The attribute is only worth having if it fires. Build a call that lies about
# its arguments and require a diagnostic. -Werror=format turns the warning
# into the failure, so a compiler that ignored the attribute compiles this and
# the case fails.
{
    echo '#include <stdio.h>'
    echo '#include <stdarg.h>'
    echo '#include <stdint.h>'
    echo 'uint8_t debug_en = 1;'
    grep 'void debug(const char \*format, \.\.\.) __attribute__' "$VIS"
    printf '%s\n' "$debug_body"
    printf '%s\n' 'int main(void){ debug("[t] %d\n", "not an integer"); return 0; }'
} > "$work/wrong.c"

if "$CXX" -x c++ -O0 -Werror=format -o "$work/wrong" "$work/wrong.c" \
        > "$work/wrong.log" 2>&1; then
    note "a call whose arguments do not match is rejected" FAIL
else
    grep -q 'format' "$work/wrong.log" \
        && note "a call whose arguments do not match is rejected" OK \
        || note "a call was rejected, but not for its format" FAIL
fi

# And the opposite, so the case above is not passing because the harness
# itself is broken.
{
    echo '#include <stdio.h>'
    echo '#include <stdarg.h>'
    echo '#include <stdint.h>'
    echo 'uint8_t debug_en = 1;'
    grep 'void debug(const char \*format, \.\.\.) __attribute__' "$VIS"
    printf '%s\n' "$debug_body"
    printf '%s\n' 'int main(void){ debug("[t] %d\n", 41); return 0; }'
} > "$work/right.c"

"$CXX" -x c++ -O0 -Werror=format -o "$work/right" "$work/right.c" \
        > "$work/right.log" 2>&1 \
    && note "a call whose arguments do match is accepted" OK \
    || { note "a call whose arguments do match is accepted" FAIL; \
         sed 's/^/    /' "$work/right.log"; }

echo "===== the shipped call sites agree with their formats ====="

# The build is what really proves this, but it needs the MaixCDK image. Count
# the calls that carry a specifier so a future reader knows how much of the
# file the attribute is checking, and fail if the file ever loses them all,
# which would mean this suite is guarding nothing.
withargs=$(grep -o 'debug("[^"]*%[^"]*"' "$VIS" | wc -l | tr -d ' ')
[ "$withargs" -ge 20 ] \
    && note "$withargs calls carry a format specifier" OK \
    || note "only $withargs calls carry a format specifier, wanted 20 or more" FAIL

echo
if [ "$fails" = 0 ]; then
    echo "test-debug-passes-arguments.sh: all cases passed"
    exit 0
fi
echo "test-debug-passes-arguments.sh: $fails case(s) failed"
exit 1
