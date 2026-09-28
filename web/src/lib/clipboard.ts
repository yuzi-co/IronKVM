// writeClipboardText copies text. The Clipboard API needs a secure context,
// and a board on plain http is not one, so it falls back to execCommand.
export async function writeClipboardText(text: string) {
  if (window.isSecureContext === true && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch {
      // Fall through to the legacy path for HTTP deployments and browser quirks.
    }
  }

  const textArea = document.createElement('textarea');
  const selection = document.getSelection();
  const selectedRange = selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null;

  textArea.value = text;
  textArea.setAttribute('readonly', '');
  textArea.style.position = 'fixed';
  textArea.style.left = '-9999px';
  textArea.style.top = '0';
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();
  textArea.setSelectionRange(0, text.length);

  try {
    if (!document.execCommand('copy')) {
      throw new Error('copy command failed');
    }
  } finally {
    document.body.removeChild(textArea);
    if (selectedRange && selection) {
      selection.removeAllRanges();
      selection.addRange(selectedRange);
    }
  }
}

export type ClipboardReadResult =
  { ok: true; text: string } | { ok: false; reason: 'unavailable' | 'denied' | 'failed' };

// readClipboardText reads the clipboard's text. Reading has no legacy path to
// fall back to: outside a secure context, which a board on plain http is, the
// browser offers no way to read the clipboard at all. The caller is told so,
// and can offer a box to paste into instead.
export async function readClipboardText(): Promise<ClipboardReadResult> {
  if (window.isSecureContext !== true || !navigator.clipboard?.readText) {
    return { ok: false, reason: 'unavailable' };
  }

  try {
    return { ok: true, text: await navigator.clipboard.readText() };
  } catch (error) {
    if (error instanceof Error && error.name === 'NotAllowedError') {
      return { ok: false, reason: 'denied' };
    }
    return { ok: false, reason: 'failed' };
  }
}
