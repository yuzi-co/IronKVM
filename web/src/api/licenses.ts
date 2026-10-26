import { http } from '@/lib/http.ts';
import { licenseFileUrl } from '@/lib/licenses.ts';
import { getBaseUrl } from '@/lib/service.ts';

// list the files of the image's license manifest. A 404 means the image has
// none (a development build) or the server predates the route.
export function listLicenses() {
  return http.get('/api/licenses');
}

// the link that opens one manifest file as plain text
export function getLicenseFileUrl(path: string) {
  return licenseFileUrl(getBaseUrl('http'), path);
}
