import { createContext, useContext } from 'react';

// SettingsNav lets a page send the user to another settings page, as the VPN
// page does for swap, which lives under Performance.
export const SettingsNav = createContext<{ openTab: (id: string) => void }>({
  openTab: () => {}
});

export function useSettingsNav() {
  return useContext(SettingsNav);
}
