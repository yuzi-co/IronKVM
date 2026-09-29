// Pure helpers for the update row of a downloaded component.

type Status = {
  installed: boolean;
  version: string;
  latest: string;
  checkedAt: string | null;
  checkError: string;
  updateAvailable: boolean;
  unverifiable: string;
  job: { state: string };
};

// compareVersions compares dotted versions number by number, each with an
// optional leading "v": -1 when a is older, 0 when equal, 1 when newer. It is
// the server's CompareVersions.
export function compareVersions(a: string, b: string): number {
  const pa = a.replace(/^v/, '').split('.');
  const pb = b.replace(/^v/, '').split('.');
  for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
    const x = parseInt(pa[i] ?? '0', 10) || 0;
    const y = parseInt(pb[i] ?? '0', 10) || 0;
    if (x !== y) return x < y ? -1 : 1;
  }
  return 0;
}

export type UpdateView =
  | 'notInstalled'
  | 'running'
  | 'unchecked'
  | 'checkFailed'
  | 'available'
  | 'unverifiable'
  | 'upToDate';

// updateView is what the row shows for a status.
export function updateView(st: Status): UpdateView {
  if (st.job.state === 'running') return 'running';
  if (!st.installed) return 'notInstalled';
  if (!st.checkedAt) return 'unchecked';
  if (st.checkError) return 'checkFailed';
  if (st.updateAvailable) return 'available';
  if (st.latest && compareVersions(st.latest, st.version) > 0) return 'unverifiable';
  return 'upToDate';
}
