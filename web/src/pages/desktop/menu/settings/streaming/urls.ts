// The addresses a player is given for the RTSP server and the SRT listener
// (ironkvm-dist#142, #143). Pure, so node --test can check them.

// The placeholders shown where the player's secrets go.
export const TOKEN_PLACEHOLDER = '<token>';
export const PASSPHRASE_PLACEHOLDER = '<passphrase>';

// The user name a player sends. The server reads only the password, so any
// name does; one is needed because a URL cannot carry a password without it.
const RTSP_USER = 'viewer';

// hostPart puts an IPv6 address in brackets, as a URL needs.
export function hostPart(host: string): string {
  return host.includes(':') && !host.startsWith('[') ? `[${host}]` : host;
}

// A placeholder is shown as it is; a real value is escaped for the URL.
function encode(value: string, placeholder: string): string {
  return value === placeholder ? value : encodeURIComponent(value);
}

// rtspUrl is the address with the credentials in it, which OBS, VLC and
// ffmpeg all take.
export function rtspUrl(
  host: string,
  port: number,
  path: string,
  token: string = TOKEN_PLACEHOLDER
): string {
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  return `rtsp://${RTSP_USER}:${encode(token, TOKEN_PLACEHOLDER)}@${hostPart(host)}:${port}${cleanPath}`;
}

// srtUrl is the address for OBS's Media Source and ffmpeg: the board
// listens, so the player calls. The latency is left to the board's setting,
// which the handshake takes when it is the larger.
export function srtUrl(
  host: string,
  port: number,
  passphrase: string = PASSPHRASE_PLACEHOLDER,
  token: string = TOKEN_PLACEHOLDER
): string {
  return (
    `srt://${hostPart(host)}:${port}?mode=caller` +
    `&passphrase=${encode(passphrase, PASSPHRASE_PLACEHOLDER)}` +
    `&streamid=${encode(token, TOKEN_PLACEHOLDER)}`
  );
}
