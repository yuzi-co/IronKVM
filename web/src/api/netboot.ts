import { http } from '@/lib/http.ts';

// Install downloads dnsmasq and the boot files, which takes minutes.
const LONG = { timeout: 10 * 60 * 1000 };

export type NetbootLease = {
  mac: string;
  ip: string;
  hostname: string;
  expires: string;
};

export type NetbootBoot = {
  time: string;
  client: string;
  what: string;
};

export type NetbootStatus = {
  usb: boolean;
  lan: boolean;
  onData: boolean;
  installed: boolean;
  version: string;
  link: {
    mode: string;
    board: string;
    host: string;
    active: boolean;
  };
  usbRunning: boolean;
  menuUrl: string;
  lanRunning: boolean;
  lanInterface: string;
  lanNetwork: string;
  lanError: string;
  images: string[] | null;
  leases: NetbootLease[] | null;
  boots: NetbootBoot[] | null;
  usbLog: string[] | null;
  lanLog: string[] | null;
};

export function getNetbootStatus() {
  return http.get('/api/netboot/status');
}

export function setNetbootSettings(usb: boolean, lan: boolean) {
  return http.post('/api/netboot/settings', { usb, lan });
}

export function installNetboot() {
  return http.post('/api/netboot/install', undefined, LONG);
}

export function uninstallNetboot() {
  return http.post('/api/netboot/uninstall');
}
