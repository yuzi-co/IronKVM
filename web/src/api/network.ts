import { http } from '@/lib/http.ts';

export type DNSMode = 'manual' | 'dhcp';
export type EthernetMode = 'static' | 'dhcp';

// wake on lan
export function wol(mac: string) {
  const data = {
    mac
  };
  return http.post('/api/network/wol', data);
}

// get wake-on-lan macs history
export function getWolMacs() {
  return http.get('/api/network/wol/mac');
}

export function deleteWolMac(mac: string) {
  const data = {
    mac
  };
  return http.delete('/api/network/wol/mac', data);
}

// set Mac name
export function setWolMacName(mac: string, name: string) {
  return http.post('/api/network/wol/mac/name', { mac, name });
}

// get wifi information
export function getWiFi() {
  return http.get('/api/network/wifi');
}

// connect wifi without auth (only available in wifi configuration mode)
export function connectWifiNoAuth(ssid: string, password: string, apPassword?: string) {
  const data = {
    ssid,
    password
  };
  return http.post('/api/network/wifi', data, {
    headers: {
      'X-AP-Key': apPassword || ''
    }
  });
}

// verify ap login
export function verifyApLogin(apPassword: string) {
  return http.post(
    '/api/network/wifi/verify',
    {},
    {
      headers: {
        'X-AP-Key': apPassword || ''
      }
    }
  );
}

// connect wifi
export function connectWifi(ssid: string, password: string) {
  const data = {
    ssid,
    password
  };
  return http.post('/api/network/wifi/connect', data);
}

// disconnect wifi
export function disconnectWifi() {
  return http.post('/api/network/wifi/disconnect');
}

export function getDNS() {
  return http.get('/api/network/dns');
}

export function setDNS(mode: DNSMode, servers: string[]) {
  return http.post('/api/network/dns', { mode, servers });
}

export function getEthernet() {
  return http.get('/api/network/ethernet');
}

// Applies the addressing on trial. The device puts the previous settings back
// unless confirmEthernet arrives with this token before the window closes, so
// a wrong address costs a wait rather than a trip to the device.
export function setEthernet(
  mode: EthernetMode,
  address: string,
  prefix: number,
  gateway: string,
  trialSeconds?: number
) {
  return http.post('/api/network/ethernet', {
    mode,
    address,
    prefix,
    gateway,
    trialSeconds: trialSeconds ?? 0
  });
}

export function confirmEthernet(token: string) {
  return http.post('/api/network/ethernet/confirm', { token });
}
