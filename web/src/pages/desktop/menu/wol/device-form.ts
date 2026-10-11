// The add/edit form of a Wake-on-LAN device. Plain functions, so node --test can
// run them; the server validates again.

export type DeviceForm = {
  oldMac: string;
  name: string;
  mac: string;
  broadcast: string;
  // 0 is the Ethernet frame the board has always sent; 7 and 9 are UDP.
  port: number;
  secureOn: string;
  // A stored password is never sent to the page: this says one exists.
  hasSecureOn: boolean;
  clearSecureOn: boolean;
  ip: string;
};

export const emptyForm: DeviceForm = {
  oldMac: '',
  name: '',
  mac: '',
  broadcast: '',
  port: 0,
  secureOn: '',
  hasSecureOn: false,
  clearSecureOn: false,
  ip: ''
};

const macRe = /^([0-9a-f]{2}[:-]?){5}[0-9a-f]{2}$|^([0-9a-f]{4}\.){2}[0-9a-f]{4}$/i;

export function isMac(value: string): boolean {
  return macRe.test(value.trim());
}

function isIPv4(value: string): boolean {
  const parts = value.split('.');
  return (
    parts.length === 4 &&
    parts.every((p) => /^\d{1,3}$/.test(p) && Number(p) >= 0 && Number(p) <= 255)
  );
}

// isBroadcast accepts an IPv4 address or an IPv4 CIDR subnet, or nothing.
export function isBroadcast(value: string): boolean {
  const v = value.trim();
  if (v === '') return true;
  const [address, prefix, ...rest] = v.split('/');
  if (rest.length > 0 || !isIPv4(address)) return false;
  if (prefix === undefined) return true;
  return /^\d{1,2}$/.test(prefix) && Number(prefix) >= 1 && Number(prefix) <= 30;
}

export function isHostIp(value: string): boolean {
  const v = value.trim();
  return v === '' || isIPv4(v);
}

export function isSecureOn(value: string): boolean {
  const v = value.trim();
  return v === '' || isMac(v);
}

export type FormErrors = Partial<Record<'mac' | 'broadcast' | 'secureOn' | 'ip', true>>;

export function validateForm(form: DeviceForm): FormErrors {
  const errors: FormErrors = {};
  if (!isMac(form.mac)) errors.mac = true;
  if (!isBroadcast(form.broadcast)) errors.broadcast = true;
  if (!isSecureOn(form.secureOn)) errors.secureOn = true;
  if (!isHostIp(form.ip)) errors.ip = true;
  return errors;
}

export function toRequest(form: DeviceForm) {
  return {
    oldMac: form.oldMac || undefined,
    mac: form.mac.trim(),
    name: form.name.trim(),
    broadcast: form.broadcast.trim(),
    port: form.port,
    secureOn: form.secureOn.trim(),
    clearSecureOn: form.clearSecureOn,
    ip: form.ip.trim()
  };
}
