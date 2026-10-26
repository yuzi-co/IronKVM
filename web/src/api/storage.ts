import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export type DriveId = 'disk' | 'cdrom';

export type Drive = {
  id: DriveId;
  type: DriveId;
  file: string;
  ro: boolean;
  // Bytes of the served file; 0 for a device. Absent from an older server.
  size?: number;
  // The served file's path is gone: deleted or renamed while in the drive.
  missing?: boolean;
};

// get image list
export function getImages() {
  return http.get('/api/storage/image');
}

// list the virtual drives: the disk, and the CD when the gadget has it
export function getDrives() {
  return http.get('/api/storage/drives');
}

// insert an image into a drive; ro applies to the disk only
export function insertDrive(id: DriveId, file: string, ro: boolean) {
  return http.post(`/api/storage/drives/${id}/insert`, { file, ro });
}

// eject a drive's image
export function ejectDrive(id: DriveId) {
  return http.post(`/api/storage/drives/${id}/eject`);
}

export function deleteImage(file: string) {
  const data = {
    file
  };
  return http.post('/api/storage/image/delete', data);
}

// The file browser: the files inside an image, named by the image's path as
// the drives name it, and a path inside the image.

export type ImageFile = {
  name: string;
  size: number;
  isDir: boolean;
  modifiedAt: string;
};

export type ImageFiles = {
  file: string;
  path: string;
  // fat12, fat16, fat32, exfat, iso9660 or udf
  fs: string;
  // Files can be written now: the filesystem can be, and no drive serves it.
  writable: boolean;
  // The drive serving the image, '' when none does.
  drive: DriveId | '';
  entries: ImageFile[];
  truncated: boolean;
  free: number;
  total: number;
};

// Error codes of the file routes the UI tells apart.
export const FILES_BUSY = -12;
export const FILES_NOT_BROWSABLE = -13;
export const FILES_NO_SPACE = -14;
export const FILES_READ_ONLY = -15;

export function listFiles(file: string, path: string) {
  return http.get('/api/storage/files', { file, path });
}

// The URL a file downloads from. The session cookie authenticates it, so a
// plain link streams the file without holding it in memory.
export function fileDownloadUrl(file: string, path: string) {
  const query = new URLSearchParams({ file, path });
  return `${getBaseUrl('http')}/api/storage/files/download?${query.toString()}`;
}

// Write a file into the directory path of an image. The image and directory go
// in the query string so the body holds nothing but the file.
export function uploadFile(
  file: string,
  path: string,
  data: File,
  onProgress: (percent: number) => void
) {
  const form = new FormData();
  form.append('file', data);

  return http.post(`/api/storage/files/upload?${new URLSearchParams({ file, path })}`, form, {
    timeout: 0,
    onUploadProgress: (event) => {
      if (event.total) onProgress(Math.round((event.loaded / event.total) * 100));
    }
  });
}

// Delete a file, or a directory with everything in it.
export function deleteFile(file: string, path: string) {
  return http.post('/api/storage/files/delete', { file, path });
}
