# benburkert/openpgp, aes/keywrap only

The RFC 3394 AES key wrap that github.com/datarhei/gosrt imports, copied
unchanged from github.com/benburkert/openpgp at commit c2471f86866c
(v0.0.0-20160410205803-c2471f86866c). The repository has no LICENSE file,
only the BSD-2-Clause header in the source, so LICENSE here repeats that
header for licence tooling. server/go.mod points the module here with a
replace directive; nothing else from the repository is used.
