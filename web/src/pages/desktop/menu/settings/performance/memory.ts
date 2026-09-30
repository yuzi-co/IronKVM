// The shape of GET /api/vm/memory. Every size is in bytes.
export type MemorySwap = {
  name: string;
  kind: 'zram' | 'file' | 'partition';
  size: number;
  used: number;
};

export type MemoryProcess = { name: string; rss: number };

// A limit is 0 when the group sets none.
export type MemoryGroup = { current: number; high: number; max: number };

export type MemoryStatus = {
  total: number;
  available: number;
  free: number;
  swaps: MemorySwap[];
  zramMemUsed: number;
  processes: MemoryProcess[];
  addons: MemoryGroup | null;
};

// Below this share of RAM available the board is short of memory: the kernel
// is reclaiming hard, and the next stream or add-on may be what it kills. On a
// board with about 180 MB for Linux that is under 27 MB.
export const LOW_MEMORY_FRACTION = 0.15;

// ramUse is what the RAM bar shows. Used is what is not available, which
// counts the cache the kernel can drop as free, the way MemAvailable does.
export function ramUse(m: Pick<MemoryStatus, 'total' | 'available'>): {
  used: number;
  percent: number;
  low: boolean;
} {
  if (m.total <= 0) return { used: 0, percent: 0, low: false };
  const available = Math.min(Math.max(m.available, 0), m.total);
  const used = m.total - available;
  return {
    used,
    percent: Math.round((used * 100) / m.total),
    low: available < m.total * LOW_MEMORY_FRACTION
  };
}

// groupUse is the add-ons' group against memory.high, where the kernel starts
// to throttle it, or memory.max without one. pressed is set once the group is
// within a tenth of memory.high, as the VPN page flags it.
export function groupUse(g: MemoryGroup): { limit: number; pressed: boolean } {
  return {
    limit: g.high || g.max,
    pressed: g.high > 0 && g.current >= g.high * 0.9
  };
}
