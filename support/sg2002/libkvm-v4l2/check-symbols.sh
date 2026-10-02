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
#
# Also fails when LIB does not record NEEDED libgcc_s.so.1 (READELF names the
# readelf to use; default readelf). NanoKVM-Server's crtbegin calls
# __register_frame_info through the PLT and relies on libkvm.so to bring
# libgcc_s in, as Sipeed's library does. Without it the server jumps to address
# 0 at start under musl. See the note on the libkvm.so rule in the Makefile.
set -eu

nm=$1
lib=$2
list=$3
vendor=${4:-}
readelf=${READELF:-readelf}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

status=0

"$readelf" -d "$lib" | sed -n 's/.*(NEEDED).*\[\(.*\)\].*/\1/p' > "$tmp/needed"
echo "NEEDED: $(tr '\n' ' ' < "$tmp/needed")"
for needed in libgcc_s.so.1 libc.so; do
	if ! grep -qx "$needed" "$tmp/needed"; then
		echo "MISSING NEEDED $needed in $lib"
		status=1
	fi
done

grep -v '^#' "$list" | sed '/^$/d' | sort -u > "$tmp/want"
"$nm" -D --defined-only "$lib" | awk '{print $2, $3}' | sort -u > "$tmp/ours_typed"
awk '$1 == "T" {print $2}' "$tmp/ours_typed" | sort -u > "$tmp/ours"
awk '{print $2}' "$tmp/ours_typed" | sort -u > "$tmp/ours_all"

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
