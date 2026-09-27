import { http } from '@/lib/http.ts';

export type DriveId = 'disk' | 'cdrom';

export type Drive = {
  id: DriveId;
  type: DriveId;
  file: string;
  ro: boolean;
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
