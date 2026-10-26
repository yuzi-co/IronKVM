// What the board reports about images served straight from a URL
// (ironkvm-dist#106), and the small rules the Media menu shows them by.

export type UrlMount = {
  drive: 'cdrom' | 'disk';
  // The nbd device the drive serves, as the drive list names it.
  device: string;
  // The URL without credentials or query.
  url: string;
  host: string;
  size: number;
  bytesRead: number;
  cache: { hits: number; misses: number; prefetched: number };
  // The last read's failure, empty once a read works again.
  error?: string;
  since: string;
};

export type UrlMediaStatus = {
  available: boolean;
  reason?: string;
  mounts: UrlMount[];
};

// isHttpUrl accepts what the board will try: an http or https URL with a host.
export function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value.trim());
    return (url.protocol === 'http:' || url.protocol === 'https:') && url.hostname !== '';
  } catch {
    return false;
  }
}

// urlMountLabel names a drive's device by the image it serves: the file name
// of the URL and its host. Anything else is not a URL mount and gets null.
export function urlMountLabel(file: string, mounts: UrlMount[]): string | null {
  const mount = mounts.find((m) => m.device === file);
  if (!mount) return null;
  const name = urlFileName(mount.url);
  return name ? `${name} (${mount.host})` : mount.host;
}

// urlFileName is the last part of the URL's path, decoded, or '' when it has
// none.
function urlFileName(url: string): string {
  try {
    return decodeURIComponent(new URL(url).pathname.replace(/^.*\//, ''));
  } catch {
    return '';
  }
}

// cacheHitRate is the share of reads the board answered from its cache, in
// whole percent, or null before the first read.
export function cacheHitRate(cache: UrlMount['cache']): number | null {
  const total = cache.hits + cache.misses;
  if (total === 0) return null;
  return Math.round((cache.hits / total) * 100);
}
