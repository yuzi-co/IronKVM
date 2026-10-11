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

// The managed device list (feature wol-devices). A wake by name sends the name
// instead of the MAC.
export interface WolDevice {
  mac: string;
  name: string;
  broadcast: string;
  port: number;
  hasSecureOn: boolean;
  ip: string;
}

export interface WolDeviceInput {
  oldMac?: string;
  mac: string;
  name: string;
  broadcast: string;
  port: number;
  secureOn: string;
  clearSecureOn: boolean;
  ip: string;
}

export interface WolNeighbor {
  mac: string;
  ip: string;
  name: string;
  listed: boolean;
}

export function getWolDevices() {
  return http.get('/api/network/wol/devices');
}

export function saveWolDevice(device: WolDeviceInput) {
  return http.post('/api/network/wol/device', device);
}

export function checkWolDevice(mac: string) {
  return http.get('/api/network/wol/check', { mac });
}

export function scanWol() {
  return http.get('/api/network/wol/scan');
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

// The collector S01syslogd forwards the system log to. target is what is
// saved, active what the running syslogd forwards to, and supported whether
// this image's init script reads the setting at all.
export type SyslogState = {
  target: string;
  active: string;
  supported: boolean;
};

export function getSyslog() {
  return http.get('/api/network/syslog');
}

// An empty target turns forwarding off.
export function setSyslog(target: string) {
  return http.post('/api/network/syslog', { target });
}

export function testSyslog() {
  return http.post('/api/network/syslog/test');
}

// The client networks allowed to reach the board. An empty list allows every
// address. client is the address the request came from, which a saved list
// must include.
export type AllowlistState = {
  networks: string[];
  client: string;
};

export function getAllowlist() {
  return http.get('/api/network/allowlist');
}

export function setAllowlist(networks: string[]) {
  return http.post('/api/network/allowlist', { networks });
}

// IPv6 on the wired interface (feature ipv6-settings). Unlike the IPv4 address
// it needs no trial: the device saves the setting after it applied, and puts
// the previous one back when applying fails.
export type IPv6Mode = 'slaac' | 'dhcpv6' | 'static' | 'off';

export type IPv6State = {
  supported: boolean;
  mode: IPv6Mode;
  address: string;
  prefix: number;
  gateway: string;
  dhcpv6Available: boolean;
  client: string;
  live: {
    interface: string;
    disabled: boolean;
    addresses: { address: string; prefix: number; scope: 'link' | 'ula' | 'global' }[];
    gateway: string;
  };
};

export function getIPv6() {
  return http.get('/api/network/ipv6');
}

export function setIPv6(mode: IPv6Mode, address: string, prefix: number, gateway: string) {
  return http.post('/api/network/ipv6', { mode, address, prefix, gateway });
}
