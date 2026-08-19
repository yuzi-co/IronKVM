#!/bin/sh
# Assert that every suite under tools/ reports its result the way run-tests.sh
# reads it.
#
#   test-suite-status.sh
#
# run-tests.sh treats the exit status of a suite as the whole contract:
#
#   0   every case passed
#   2   the suite cannot run here, and it says why on the last line
#   1   a case failed, and that is a defect in what the suite tests
#
# A suite that exits its failure count breaks the contract at exactly one
# value. Two failing cases exit 2, and the sweep prints SKIP. The reader is
# told the suite did not run, when the suite ran and found two defects. The
# status that means "nothing was proved here" is the status a real defect
# arrives as, which is the worst way for a test harness to be wrong.
#
# The check is static, because no caller can drive an arbitrary suite into two
# failing cases from outside. It works from the suite's own code instead: it
# finds every variable a suite increments, which is how each of them counts its
# failures, and it requires that no exit statement names one. The rule is
# derived from the suite rather than written down here, so a counter under a
# new name is caught as well.
cd "$(dirname "$0")/.." || exit 1

fails=0
note() { printf '  %-56s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# hwtest-*.sh is included although run-tests.sh sweeps test-*.sh only. A person
# runs those by hand and reads the same status, and the contract costs nothing
# to keep whole.
suites=$(find tools -name 'test-*.sh' -o -name 'hwtest-*.sh' | sort)

echo "===== the suite list is not empty ====="

# Without this the whole file passes on a checkout where the enumeration finds
# nothing, which is the failure the run-tests.sh status contract exists to stop
# elsewhere. A check that reports OK having inspected zero files is worse than
# no check.
count=$(echo "$suites" | grep -c '^tools/')
if [ "$count" -ge 40 ]; then
    note "$count suites found" OK
else
    note "only $count suites found, so the enumeration is broken" FAIL
fi

echo
echo "===== no suite exits the number of cases that failed ====="

# A counter is a variable a suite assigns from itself plus something:
#
#     fails=$((fails + 1))
#
# The name is not assumed. Every suite here calls it "fails" today, and the
# next one need not.
counters_of() {
    grep -oE '\b[A-Za-z_][A-Za-z0-9_]*=\$\(\([A-Za-z_][A-Za-z0-9_]*[[:space:]]*\+' "$1" \
        | sed 's/=.*//' | sort -u
}

bad=
for s in $suites; do
    for c in $(counters_of "$s"); do
        # "exit $c", "exit \"$c\"" and "exit \"${c}\"" all mean the same thing.
        if grep -qE "\bexit[[:space:]]+\"?\\$\{?$c\b" "$s"; then
            bad="$bad $s"
            break
        fi
    done
done

if [ -z "$bad" ]; then
    note "every suite exits a status, not a count" OK
else
    for s in $bad; do
        note "$s exits its failure count" FAIL
    done
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== every suite reports a status run-tests.sh can read ====="
else
    echo "===== $fails check(s) FAILED ====="
    exit 1
fi
