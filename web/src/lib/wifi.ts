// wifiCredentialsError checks a network name and password the way the server
// does (ConnectWifiReq in server/proto/network.go). An empty password joins an
// open network; anything else is a WPA passphrase of 8 to 63 characters.
export function wifiCredentialsError(ssid: string, password: string): '' | 'ssid' | 'password' {
  if (ssid === '' || ssid.length > 32) return 'ssid';
  if (password !== '' && (password.length < 8 || password.length > 63)) return 'password';
  return '';
}
