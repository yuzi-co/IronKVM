#!/bin/sh
# Check the exported symbols of the built libkvm.so.
#
#   check-symbols.sh NM LIB SYMBOL_LIST [VENDOR_LIB]
#
# Fails when LIB does not define every symbol in SYMBOL_LIST as a strong text
# symbol (nm type T). A weak definition would still resolve, but the vendor
# library's are strong and the server was built against those. Also lists any
# other symbol LIB exports, and, given the vendor library, confirms it exports
# the same functions, so the list cannot drift from what the server had.
set -eu

nm=$1
lib=$2
list=$3
vendor=${4:-}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

grep -v '^#' "$list" | sed '/^$/d' | sort -u > "$tmp/want"
"$nm" -D --defined-only "$lib" | awk '{print $2, $3}' | sort -u > "$tmp/ours_typed"
awk '$1 == "T" {print $2}' "$tmp/ours_typed" | sort -u > "$tmp/ours"
awk '{print $2}' "$tmp/ours_typed" | sort -u > "$tmp/ours_all"

status=0
missing=$(comm -23 "$tmp/want" "$tmp/ours")
if [ -n "$missing" ]; then
	echo "MISSING from $lib (as T):"
	echo "$missing" | sed 's/^/  /'
	status=1
fi

extra=$(comm -13 "$tmp/want" "$tmp/ours_all" | grep -vxE '_init|_fini' || true)
if [ -n "$extra" ]; then
	echo "other exported symbols:"
	echo "$extra" | sed 's/^/  /'
fi

if [ -n "$vendor" ] && [ -f "$vendor" ]; then
	"$nm" -D --defined-only "$vendor" | awk '$2 == "T" {print $3}' | sort -u > "$tmp/vendor"
	not_in_vendor=$(comm -23 "$tmp/want" "$tmp/vendor")
	if [ -n "$not_in_vendor" ]; then
		echo "listed but not exported by the vendor library:"
		echo "$not_in_vendor" | sed 's/^/  /'
		status=1
	else
		echo "vendor library exports all $(wc -l < "$tmp/want") listed symbols too"
	fi
fi

if [ "$status" -eq 0 ]; then
	echo "OK: $lib exports all $(wc -l < "$tmp/want") symbols in $list"
fi
exit "$status"
