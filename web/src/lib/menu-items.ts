// The menu bar entries a user can hide, as stored in the browser.
//
// The Image and Download buttons became one Media button. A saved list that
// still names either of them is rewritten once: Media is hidden only when both
// of the old buttons were, since hiding one of them never hid the other's job.
// Scripts, Wake on LAN and PicoClaw kept their ids and are now hidden one by
// one inside the Tools menu, so they need no rewrite.

const LEGACY_MEDIA_ITEMS = ['image', 'download'];

// migrateMenuDisabledItems returns the list in today's ids. It returns the
// same array when there is nothing to change, so a caller can tell whether the
// stored copy needs writing back.
export function migrateMenuDisabledItems(items: string[]): string[] {
  if (!items.some((item) => LEGACY_MEDIA_ITEMS.includes(item))) return items;

  const bothHidden = LEGACY_MEDIA_ITEMS.every((item) => items.includes(item));
  const rest = items.filter((item) => !LEGACY_MEDIA_ITEMS.includes(item));
  if (bothHidden && !rest.includes('media')) rest.push('media');

  return rest;
}
