#!/bin/sh
# Exercise the gadget's endpoint budget, taken straight out of the script that
# ships so the test cannot drift from it.
#
#   test-usb-endpoints.sh [path-to-S03usbdev]
#
# Not destructive: no gadget is built and no marker is written.
SV=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S03usbdev}
[ -f "$SV" ] || { echo "usage: test-usb-endpoints.sh <S03usbdev>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

sed -n '/^# --- endpoint budget ---$/,/^# --- end endpoint budget ---$/p' "$SV" > "$WORK/budget.sh"
[ -s "$WORK/budget.sh" ] || { echo "could not extract the endpoint budget block"; exit 1; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== what each function costs, in each direction ====="
cost_case() {
    got=$(WORK="$WORK" sh -c ". \"\$WORK/budget.sh\"; usb_in_cost $1")
    [ "$got" = "$2" ] && note "$1 costs $got in" OK || note "$1 costs $got in, want $2" FAIL
    got=$(WORK="$WORK" sh -c ". \"\$WORK/budget.sh\"; usb_out_cost $1")
    [ "$got" = "$3" ] && note "$1 costs $got out" OK || note "$1 costs $got out, want $3" FAIL
}
#         name      in out
cost_case console    2  1
cost_case network    2  1
cost_case disk       1  1
cost_case audio      0  1
cost_case nonsense   0  0

echo
echo "===== the totals, against six in and seven out ====="
# HID is three inbound of the six before anything optional is added. The
# speaker costs nothing inbound, which is why it fits where the network does
# not.
used_case() {
    desc="$1"; set="$2"; want_in="$3"; want_out="$4"
    got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_used_in "'"$set"'"')
    [ "$got" = "$want_in" ] && note "$desc -> $got in" OK || note "$desc -> $got in, want $want_in" FAIL
    got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_used_out "'"$set"'"')
    [ "$got" = "$want_out" ] && note "$desc -> $got out" OK || note "$desc -> $got out, want $want_out" FAIL
}
#                                                                  in out
used_case "hid alone"              ""                               3  3
used_case "hid + console"          "console"                        5  4
used_case "hid + console + audio"  "console audio"                  5  5
used_case "hid + console + disk"   "console disk"                   6  5
used_case "hid + disk + network"   "disk network"                   6  5
used_case "hid + console + net"    "console network"                7  5
used_case "everything"             "console disk network audio"     8  7

echo
echo "===== the hid cost, unstubbed ====="
# Every case above stubs the HID cost so the arithmetic can pick it directly.
# That leaves the real functions untested: a mutation that charges HID nothing
# would let the console and the network through together, the gadget would ask
# for seven inbound endpoints of six, and one interrupt IN endpoint would fail
# to come up with nothing said about it.
hid_cost_case() {
    desc="$1"; marker="$2"; fn="$3"; want="$4"
    got=$(WORK="$WORK" MARKER="$marker" FN="$fn" sh -c '
        . "$WORK/budget.sh"
        BOOT="$WORK/boot-hidcost-$$"; mkdir -p "$BOOT"
        [ -n "$MARKER" ] && : > "$BOOT/$MARKER"
        usb_marker() { [ -e "$BOOT/$1" ]; }
        "$FN"
    ')
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}
hid_cost_case "hid built by default, in"    ""          usb_hid_in_cost   3
hid_cost_case "hid built by default, out"   ""          usb_hid_out_cost  3
hid_cost_case "disable_hid present, in"     disable_hid usb_hid_in_cost   0
hid_cost_case "disable_hid present, out"    disable_hid usb_hid_out_cost  0

echo
echo "===== resolving a set that does not fit ====="
# Audio goes first, then the network. The console outranks both because it is
# the only way into a board whose network is gone.
resolve_case() {
    desc="$1"; set="$2"; want="$3"
    got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_resolve "'"$set"'" 6 7')
    [ "$got" = "$want" ] && note "$desc -> [$got]" OK || note "$desc -> [$got], want [$want]" FAIL
}
# Everything enabled keeps three of the four. Giving up the lowest priority
# members instead would settle on console+disk and leave an endpoint unused,
# losing audio for nothing.
resolve_case "everything keeps console + disk + audio" "console disk network audio" "console disk audio"
resolve_case "console + audio already fits"            "console audio"              "console audio"
resolve_case "console + disk + audio is exactly 6 in"  "console disk audio"         "console disk audio"
resolve_case "disk + network + audio is exactly 6 in"  "disk network audio"         "disk network audio"
# The pair the flat budget of nine got wrong. The console and the network take
# two inbound endpoints each, HID takes three, and the controller has six. The
# speaker takes the place the network cannot have, because it costs nothing
# inbound at all.
resolve_case "console + network does not fit"          "console network"            "console"
resolve_case "the speaker fits where the network does not" "console network audio"  "console audio"
resolve_case "nothing enabled stays nothing"           ""                           ""

# The output is in priority order regardless of how the set arrived, because
# configfs numbers interfaces in link order and the caller links what this
# prints.
resolve_case "the result is ordered, not as given"     "audio disk console"         "console disk audio"

# A lower priority function must never take a place from a higher one. The
# console costs 3 and audio costs 1, so a resolve that filled cheaply first
# would keep audio and drop the console - the exact inversion that leaves a
# board with no way in when its network dies.
got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_resolve "console network audio" 6 7')
case " $got " in
    *" console "*) note "the console outranks audio for the last place" OK ;;
    *)             note "the console lost its place to a cheaper function" FAIL ;;
esac

echo
echo "===== the two orderings are one ranking ====="
keep=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_keep_order')
drop=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_drop_order')
reversed=$(for name in $keep; do echo "$name"; done | sed '1!G;h;$!d' | tr '\n' ' ')
[ "$(echo $reversed)" = "$(echo $drop)" ] \
    && note "drop order is the keep order reversed" OK \
    || note "keep [$keep] reversed is [$reversed], drop says [$drop]" FAIL

echo
echo "===== HID is never a candidate ====="
# The one rule with no exception. A board that gives up HID has given up being
# a KVM, so no combination of markers may reach that state.
for set in "console disk network audio" "console network" "disk network audio"; do
    got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_resolve "'"$set"'" 6 7')
    case "$got" in
        *hid*) note "resolving [$set] dropped hid" FAIL ;;
        *)     note "resolving [$set] keeps hid" OK ;;
    esac
done

# A budget so small that nothing optional fits must still not touch HID, and
# must not loop forever trying.
got=$(WORK="$WORK" HID=3 sh -c '. "$WORK/budget.sh"; usb_hid_in_cost() { echo "$HID"; }; usb_hid_out_cost() { echo "$HID"; }; usb_resolve "console disk network audio" 3 3')
[ -z "$got" ] && note "a budget of 3 leaves only hid, and terminates" OK \
              || note "a budget of 3 left [$got]" FAIL

echo
echo "===== what was given up is reported ====="
got=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_dropped "console disk network audio" "console disk"')
want="network audio"
got=$(echo $got)
[ "$got" = "$want" ] && note "dropped [$got]" OK || note "dropped [$got], want [$want]" FAIL

got=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_dropped "console audio" "console audio"')
[ -z "$got" ] && note "nothing dropped when everything fits" OK || note "dropped [$got], want nothing" FAIL

echo
echo "===== what the prune gives back ====="
# start_usb_dev only ever adds, and `stop` leaves configs/c.1 untouched, so a
# function linked by an earlier start survives a later start that dropped it.
# usb_prune_list names the config symlinks that have to go before the link pass.
prune_case() {
    desc="$1"; keep="$2"; want="$3"
    got=$(WORK="$WORK" KEEP="$keep" sh -c '. "$WORK/budget.sh"; usb_prune_list "$KEEP"')
    got=$(echo $got)
    [ "$got" = "$want" ] && note "$desc -> [$got]" OK || note "$desc -> [$got], want [$want]" FAIL
}
prune_case "an empty keep set drops all four"          "" \
    "acm.GS0 mass_storage.disk0 ncm.usb0 rndis.usb0 uac1.usb0"
prune_case "a full keep set drops none"                "console disk network audio" ""
prune_case "console + network drops disk and audio"    "console network" \
    "mass_storage.disk0 uac1.usb0"
prune_case "the boot set drops both network flavours"  "console disk audio" \
    "ncm.usb0 rndis.usb0"
# A name the table does not know keeps nothing, so everything is pruned. The
# keep set is produced by usb_resolve, but a typo there must fail closed - link
# nothing extra - rather than leave a dropped function linked.
prune_case "an unknown keep name keeps nothing"        "nonsense" \
    "acm.GS0 mass_storage.disk0 ncm.usb0 rndis.usb0 uac1.usb0"

# The rule with no exception, restated against the prune. hid.GS0, hid.GS1 and
# hid.GS2 are gated on /boot/disable_hid alone and never enter the keep set, so
# a prune that walked configs/c.1 would delete the keyboard and both mice from
# any board that is over budget. It must also print bare directory names: a
# name carrying a path lets `rm -f configs/c.1/$name` escape the config
# directory and reach functions/, whose removal blocks forever on the getty.
prune_hid=""
prune_path=""
for a in "" console
do
    for b in "" disk
    do
        for c in "" network
        do
            for d in "" audio
            do
                got=$(WORK="$WORK" KEEP="$a $b $c $d" sh -c '. "$WORK/budget.sh"; usb_prune_list "$KEEP"')
                case "$got" in *hid*) prune_hid="$prune_hid [$a $b $c $d]" ;; esac
                case "$got" in */*)   prune_path="$prune_path [$a $b $c $d]" ;; esac
            done
        done
    done
done
[ -z "$prune_hid" ] && note "no keep set makes the prune name a hid function" OK \
                    || note "the prune named hid for$prune_hid" FAIL
[ -z "$prune_path" ] && note "the prune names directories, never paths" OK \
                     || note "the prune printed a path for$prune_path" FAIL

# usb_gadget_dirs is the mapping the prune reads. HID is not in it, and neither
# is anything else the keep set cannot name.
for name in hid hid.GS0 disable_hid ""
do
    got=$(WORK="$WORK" NAME="$name" sh -c '. "$WORK/budget.sh"; usb_gadget_dirs "$NAME"')
    got=$(echo $got)
    [ -z "$got" ] && note "usb_gadget_dirs [$name] maps to nothing" OK \
                  || note "usb_gadget_dirs [$name] gave [$got], want nothing" FAIL
done

echo
echo "===== the config a second start leaves behind ====="
# The scenario the prune exists for, driven through the real functions.
#
# A toggle that switches a function off frees endpoints and can only ever let
# more in, so it cannot leave a stale link. Switching HID back on is the case
# that can: it takes three inbound endpoints that were free a moment ago, and
# nothing unlinks the function that has to give up its place.
#
# Boot with /boot/disable_hid present and usb.acm, usb.ncm, usb.disk0 and
# usb.uac all set: five inbound endpoints are wanted of six, so all four are
# kept and linked. The operator then switches HID on in the web UI, which
# removes the marker and restarts this script. The resolve now keeps
# console+disk+audio, and ncm.usb0, linked by the first start, has to be gone
# before the link pass. Left behind it makes hid(3) + acm(2) + ncm(2) + msc(1)
# = 8 inbound of 6: the gadget still binds and the UDC still reports
# "configured", the host's set-configuration fails to enable one interrupt IN
# endpoint, and every command here still exits 0 so the UI reports success.
sim=$(WORK="$WORK" sh -c '
    . "$WORK/budget.sh"

    G="$WORK/sim"; rm -rf "$G"; mkdir -p "$G/functions" "$G/configs/c.1"
    cd "$G"

    # An entry in configs/c.1 stands in for the symlink the real script makes.
    # A plain file is deliberate: rm -f treats a symlink and a file alike, and
    # a real symlink would make this case depend on the host filesystem rather
    # than on the prune. What is under test is which entries survive.
    link() { mkdir -p "functions/$1"; : > "configs/c.1/$1"; }
    listing() { ls configs/c.1 | sort | tr "\n" " "; }

    # Step 1: the first boot, with HID switched off. Nothing links hid.GS*.
    HID=0
    usb_hid_in_cost() { echo "$HID"; }
    usb_hid_out_cost() { echo "$HID"; }
    keep1=$(usb_resolve "console disk network audio" "$(usb_in_budget)" "$(usb_out_budget)")
    for name in $keep1
    do
        for dir in $(usb_gadget_dirs "$name")
        do
            [ "$dir" = rndis.usb0 ] && continue
            link "$dir"
        done
    done
    echo "KEEP1:$(echo $keep1)"
    echo "STEP1:$(listing)"

    # Step 2: the HID toggle removes /boot/disable_hid. It unlinks nothing.
    HID=3

    # Step 3: the same markers, three fewer endpoints to spend.
    keep2=$(usb_resolve "console disk network audio" "$(usb_in_budget)" "$(usb_out_budget)")
    echo "KEEP2:$(echo $keep2)"

    # Step 4: the prune, then the link pass - the order start_usb_dev uses.
    for usb_stale in $(usb_prune_list "$keep2")
    do
        rm -f "configs/c.1/$usb_stale"
    done
    link hid.GS0; link hid.GS1; link hid.GS2
    for name in $keep2
    do
        for dir in $(usb_gadget_dirs "$name")
        do
            [ "$dir" = rndis.usb0 ] && continue
            link "$dir"
        done
    done
    echo "STEP4:$(listing)"
')

sim_field() { printf '%s\n' "$sim" | sed -n "s/^$1://p" | sed 's/ *$//'; }

got=$(sim_field KEEP1)
[ "$got" = "console disk network audio" ] && note "step 1 keeps [$got]" OK \
    || note "step 1 keeps [$got], want [console disk network audio]" FAIL

got=$(sim_field STEP1)
want="acm.GS0 mass_storage.disk0 ncm.usb0 uac1.usb0"
[ "$got" = "$want" ] && note "step 1 links [$got]" OK \
                     || note "step 1 links [$got], want [$want]" FAIL

got=$(sim_field KEEP2)
[ "$got" = "console disk audio" ] && note "step 3 keeps [$got]" OK \
                                 || note "step 3 keeps [$got], want [console disk audio]" FAIL

got=$(sim_field STEP4)
want="acm.GS0 hid.GS0 hid.GS1 hid.GS2 mass_storage.disk0 uac1.usb0"
[ "$got" = "$want" ] && note "step 4 links [$got]" OK \
                     || note "step 4 links [$got], want [$want]" FAIL

case " $(sim_field STEP4) " in
    *" ncm.usb0 "*) note "the dropped network is still linked after step 4" FAIL ;;
    *)              note "the dropped network is gone after step 4" OK ;;
esac

missing=""
for dir in hid.GS0 hid.GS1 hid.GS2
do
    case " $(sim_field STEP4) " in
        *" $dir "*) ;;
        *) missing="$missing $dir" ;;
    esac
done
[ -z "$missing" ] && note "all three hid functions survive the prune" OK \
                  || note "the prune removed$missing" FAIL

echo
echo "===== the budgets are the controller's numbers, not dmesg's ====="
# dwc2 announces "EPs: 8" and that is neither budget. The controller reports
# seven endpoint pairs and six dedicated transmit FIFOs, so the two limits are
# six inbound and seven outbound. The guard must not consult dmesg at all.
got=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_in_budget')
[ "$got" = 6 ] && note "the inbound budget is 6" OK || note "the inbound budget is [$got], want 6" FAIL

got=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_out_budget')
[ "$got" = 7 ] && note "the outbound budget is 7" OK || note "the outbound budget is [$got], want 7" FAIL

# The two are not the same number. A mutation that made them equal would let
# the console and the network through together.
[ "$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_in_budget')" \
    != "$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; usb_out_budget')" ] \
    && note "the two directions have different limits" OK \
    || note "both directions report the same limit" FAIL

if grep -q dmesg "$WORK/budget.sh"; then
    note "the budget consults dmesg, which reports the wrong number" FAIL
else
    note "the budget never consults dmesg" OK
fi

# A board that reports something else must not change the answer, because the
# answer was read from the controller and the kernel's line disagrees with it.
got=$(WORK="$WORK" sh -c '. "$WORK/budget.sh"; dmesg() { echo "dwc2 4340000.usb: EPs: 8, dedicated fifos"; }; usb_in_budget')
[ "$got" = 6 ] && note "a dmesg line saying 8 does not move the budget" OK \
               || note "dmesg moved the budget to [$got]" FAIL

echo
echo "===== the network is one function, whichever marker names it ====="
# usb.ncm and usb.rndis0 are alternatives. Counting both would reserve three
# endpoints nothing uses, and the guard would refuse a function that fits.
got=$(WORK="$WORK" sh -c '
    . "$WORK/budget.sh"
    BOOT="$WORK/boot"; mkdir -p "$BOOT"
    : > "$BOOT/usb.ncm"; : > "$BOOT/usb.rndis0"
    usb_marker() { [ -e "$BOOT/$1" ]; }
    usb_enabled
')
got=$(echo $got)
[ "$got" = "network" ] && note "two network markers name one function" OK \
                       || note "gave [$got], want [network]" FAIL

echo
echo "===== usb_enabled reads the marker each function actually ships with ====="
# Only the network pair was exercised above. A typo in one of the other three
# markers (usb.uac -> usb.uac1, usb.disk0 -> usb.disk, usb.acm -> usb.acm2)
# would make usb_enabled silently under-report what is on, usb_resolve would
# then approve an over-budget set because it never sees the function it
# missed, and the gadget would refuse to bind.
enabled_n=0
enabled_case() {
    marker="$1"; want="$2"
    enabled_n=$((enabled_n + 1))
    got=$(WORK="$WORK" N="$enabled_n" MARKER="$marker" sh -c '
        . "$WORK/budget.sh"
        BOOT="$WORK/boot-enabled-$N"; mkdir -p "$BOOT"
        : > "$BOOT/$MARKER"
        usb_marker() { [ -e "$BOOT/$1" ]; }
        usb_enabled
    ')
    got=$(echo $got)
    [ "$got" = "$want" ] && note "$marker enables [$want]" OK || note "$marker gave [$got], want [$want]" FAIL
}
enabled_case usb.acm   console
enabled_case usb.disk0 disk
enabled_case usb.uac   audio

echo
echo "===== the script still parses ====="
sh -n "$SV" && note "S03usbdev is valid shell" OK || note "S03usbdev does not parse" FAIL

# A wiring check. Every case above tests the block in isolation, so the block
# could be correct and never called.
grep -q 'usb_keep=\$(usb_resolve ' "$SV" \
    && note "start_usb_dev resolves the enabled set" OK \
    || note "the budget block is never called" FAIL

# Each function must be gated on usb_kept, not reconstructed from a marker
# test that the resolve step never touches. Anchored to the exact indentation
# and the exact end of the line, so a commented-out gate - which a plain
# substring grep would still count as present - reports as missing.
gate_case() {
    desc="$1"; pattern="$2"
    grep -qE "$pattern" "$SV" \
        && note "$desc is gated on usb_kept" OK \
        || note "$desc still tests its marker directly" FAIL
}
gate_case "the console"        '^    if usb_kept console$'
gate_case "the disk"           '^    if usb_kept disk$'
gate_case "ncm"                '^    if usb_kept network && \[ -e /boot/usb\.ncm \]$'
gate_case "rndis0"             '^        if usb_kept network && \[ -e /boot/usb\.rndis0 \]$'
gate_case "the os_desc block"  '^    if usb_kept network$'
gate_case "audio"              '^    if usb_kept audio$'

# The one gate that must never appear. HID is the one function with no
# exception - a "usb_kept hid" here would always be false, because hid never
# appears in usb_keep_order, and every boot would come up without a keyboard.
hid_block=$(sed -n '/^    if \[ ! -e \/boot\/disable_hid \]$/,/^    fi$/p' "$SV")
if [ -n "$hid_block" ] && ! printf '%s\n' "$hid_block" | grep -q usb_kept; then
    note "HID is gated only on its own marker, never on usb_kept" OK
else
    note "HID's gate changed, or now depends on usb_kept" FAIL
fi

# The prune has to be wired in, and it has to run before the link pass - a
# prune that ran afterwards would remove the links that same start just made.
# Anchored to the exact indentation and the exact end of the line, so a
# commented-out loop reports as missing.
grep -qE '^    for usb_stale in \$\(usb_prune_list "\$usb_keep"\)$' "$SV" \
    && note "start_usb_dev prunes what the budget dropped" OK \
    || note "nothing prunes - a dropped function stays linked from an earlier start" FAIL

grep -qE '^        rm -f "configs/c\.1/\$usb_stale"$' "$SV" \
    && note "the prune unlinks configs/c.1 entries with rm -f" OK \
    || note "the prune does not rm -f configs/c.1/\$usb_stale" FAIL

prune_line=$(grep -nE '^        rm -f "configs/c\.1/\$usb_stale"$' "$SV" | head -n 1 | cut -d: -f1)
link_line=$(grep -nE '^ *ln -s functions/' "$SV" | head -n 1 | cut -d: -f1)
if [ -n "$prune_line" ] && [ -n "$link_line" ] && [ "$prune_line" -lt "$link_line" ]
then
    note "the prune runs before the first link" OK
else
    note "the prune is at line [$prune_line], the first link at [$link_line]" FAIL
fi

# The prune loop itself must never reach a function directory. `rmdir
# functions/acm.GS0` blocks forever on the getty that /etc/inittab respawns on
# /dev/ttyGS0, and recovery needs a full teardown of the gadget.
prune_block=$(sed -n '/^    for usb_stale in /,/^    done$/p' "$SV")
if [ -n "$prune_block" ] && ! printf '%s\n' "$prune_block" | grep -qE 'rmdir|functions/'
then
    note "the prune removes symlinks only, never a function directory" OK
else
    note "the prune touches functions/ or runs rmdir" FAIL
fi

# Dropping must not delete the operator's marker: intent has to survive so the
# function returns on its own once something else is switched off.
if sed -n '/^# --- endpoint budget ---$/,/^# --- end endpoint budget ---$/p' "$SV" | grep -qE '\brm\b|\bunlink\b'; then
    note "the budget block removes a file" FAIL
else
    note "the budget block never removes a marker" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== all endpoint cases pass ====="
else
    echo "===== $fails case(s) failed ====="
    exit 1
fi
