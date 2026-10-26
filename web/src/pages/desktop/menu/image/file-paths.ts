import type { Drive } from '@/api/storage.ts';

// Paths inside an image, as the file routes take them: slash separated,
// relative to the image root, and "/" for the root itself.

export function joinPath(dir: string, name: string): string {
  return dir === '/' || dir === '' ? name : `${dir}/${name}`;
}

export function parentPath(dir: string): string {
  const i = dir.lastIndexOf('/');
  return i <= 0 ? '/' : dir.slice(0, i);
}

export type Crumb = { name: string; path: string };

// breadcrumbs splits a directory into the steps from the root to it. The root
// has no name; the caller labels it.
export function breadcrumbs(dir: string): Crumb[] {
  const crumbs: Crumb[] = [{ name: '', path: '/' }];
  if (dir === '/' || dir === '') return crumbs;

  let path = '';
  for (const name of dir.split('/').filter(Boolean)) {
    path = path ? `${path}/${name}` : name;
    crumbs.push({ name, path });
  }
  return crumbs;
}

// What the host can do with an image decides what the browser may:
// - free: no drive serves it, so it can be browsed and changed;
// - hostReads: a drive serves it read-only, so it can be browsed only;
// - hostWrites: a drive serves it writable, so the board keeps away from it.
// The Ventoy disk reads its images read-only, so its images are hostReads.
export type BrowseState = 'free' | 'hostReads' | 'hostWrites';

export function browseState(image: string, drives: Drive[], ventoyImages: string[]): BrowseState {
  const drive = drives.find((d) => d.file === image);
  if (drive) return drive.ro ? 'hostReads' : 'hostWrites';
  return ventoyImages.includes(image) ? 'hostReads' : 'free';
}

// fileSize names a file's size in the unit that keeps it short: a listing
// holds text files of a few bytes as well as images of gigabytes.
export function fileSize(bytes: number): string {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return unit === 0 ? `${value} B` : `${value.toFixed(value < 10 ? 1 : 0)} ${units[unit]}`;
}
