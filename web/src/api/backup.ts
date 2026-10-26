import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export type Change = { field: string; from: unknown; to: unknown };

export type SectionStatus = 'change' | 'same' | 'invalid' | 'unknown';

export type SectionPreview = {
  section: string;
  status: SectionStatus;
  changes?: Change[];
  error?: string;
  warnings?: string[];
};

export type Preview = {
  version: number;
  exportedAt?: string;
  source: { application?: string; hostname?: string };
  sections: SectionPreview[];
};

export type ImportResult = { applied: string[]; rebootRequired: boolean };

export type Download = { blob: Blob; fileName: string | null };

// fetchDownload asks for a file. The server answers a failure with its usual
// JSON envelope and status 200, so a JSON reply is read as that failure.
async function fetchDownload(path: string): Promise<Download> {
  const rsp = await fetch(`${getBaseUrl('http')}${path}`, { credentials: 'include' });
  if (!rsp.ok) throw new Error(`download failed: ${rsp.status}`);

  const type = rsp.headers.get('Content-Type') ?? '';
  const disposition = rsp.headers.get('Content-Disposition');
  if (!disposition && type.startsWith('application/json')) {
    const body = (await rsp.json()) as { code: number; msg?: string };
    throw body;
  }
  return { blob: await rsp.blob(), fileName: disposition };
}

export function exportSettings() {
  return fetchDownload('/api/settings/export');
}

export function downloadDiagnostics() {
  return fetchDownload('/api/diagnostics/bundle');
}

export function previewSettings(document: unknown) {
  return http.post('/api/settings/preview', { document }) as Promise<{
    code: number;
    msg: string;
    data: Preview;
  }>;
}

export function importSettings(document: unknown, sections: string[]) {
  return http.post('/api/settings/import', { document, sections }) as Promise<{
    code: number;
    msg: string;
    data: ImportResult;
  }>;
}
