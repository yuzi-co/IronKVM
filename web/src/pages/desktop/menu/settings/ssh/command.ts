// sshCommand is what to type to log in as root. The port is spelled out only
// when it is not the default, so the common case stays short. An IPv6 address
// needs no brackets here: ssh splits user@host at the @, not at a colon.
export function sshCommand(host: string, port: number): string {
  const target = `root@${host}`;
  return port && port !== 22 ? `ssh -p ${port} ${target}` : `ssh ${target}`;
}

// A key type as ssh-keygen shows it: ED25519 for ssh-ed25519, ECDSA for any
// ecdsa-sha2-*, and the rest upper-cased without the ssh- prefix.
export function keyLabel(type: string): string {
  if (type.startsWith('ecdsa-')) return 'ECDSA';
  if (type.startsWith('sk-ecdsa-')) return 'ECDSA-SK';
  if (type.startsWith('sk-ssh-ed25519')) return 'ED25519-SK';
  const base = type.replace(/^ssh-/, '').replace(/-cert-v01@openssh\.com$/, '-CERT');
  return base.toUpperCase();
}
