import { useHidMode } from '@/hooks/useHidMode.ts';

// The touch screen exists only when the USB gadget declares it, which normal
// mode does with /boot/usb.touch and hid-only mode never does. The server
// reads that from the gadget. With /boot/disable_hid nothing is linked, so
// touch is not offered either.
export function useTouchAvailable(): boolean {
  const hidMode = useHidMode();

  return !!hidMode && hidMode.mode === 'normal' && !hidMode.hidDisabled && hidMode.touch;
}
