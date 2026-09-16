import { useCallback, useEffect, useRef, useState } from 'react';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';

import {
  getKeyboardLedStatusVisible,
  getMenuDisabledItems,
  getMenuDisplayMode
} from '@/lib/localstorage.ts';
import {
  keyboardLedStatusVisibleAtom,
  menuDisabledItemsAtom,
  menuDisplayModeAtom,
  submenuOpenCountAtom
} from '@/jotai/settings.ts';

const HIDE_TIMEOUT = 5000;

export interface MenuVisibilityState {
  isInitialized: boolean;
  isMenuExpanded: boolean;
  isMenuHidden: boolean;
  isMenuMoved: boolean;
  handleHovered: (hovered: boolean) => void;
  handleMoved: () => void;
  setIsMenuExpanded: (expanded: boolean) => void;
}

export function useMenuVisibility(): MenuVisibilityState {
  const [menuDisplayMode, setMenuDisplayMode] = useAtom(menuDisplayModeAtom);
  const setMenuDisabledItems = useSetAtom(menuDisabledItemsAtom);
  const setKeyboardLedStatusVisible = useSetAtom(keyboardLedStatusVisibleAtom);
  const submenuOpenCount = useAtomValue(submenuOpenCountAtom);

  const [isMenuExpanded, setIsMenuExpanded] = useState(() => getMenuDisplayMode() !== 'off');
  const [isMenuMoved, setIsMenuMoved] = useState(false);
  const [isMenuHovered, setIsMenuHovered] = useState(false);
  const [isMenuHidden, setIsMenuHidden] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);

  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startCountdown = useCallback(() => {
    if (menuDisplayMode !== 'auto' || submenuOpenCount > 0 || !isMenuExpanded || isMenuMoved) {
      return;
    }

    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }

    timerRef.current = setTimeout(() => {
      setIsMenuHidden(true);
    }, HIDE_TIMEOUT);
  }, [menuDisplayMode, submenuOpenCount, isMenuExpanded, isMenuMoved]);

  const stopCountdown = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Initialize menu settings
  useEffect(() => {
    const displayMode = getMenuDisplayMode();
    setMenuDisplayMode(displayMode);

    const items = getMenuDisabledItems();
    setMenuDisabledItems(items);

    setKeyboardLedStatusVisible(getKeyboardLedStatusVisible());

    // The menu fades in once the settings above have reached the atoms. The
    // flag is raised on the next frame rather than in this effect, which is
    // when that render has happened.
    const frame = requestAnimationFrame(() => setIsInitialized(true));

    return () => {
      cancelAnimationFrame(frame);
      stopCountdown();
    };
  }, [setMenuDisplayMode, setMenuDisabledItems, setKeyboardLedStatusVisible, stopCountdown]);

  // Any change to what the countdown depends on shows the menu again. That is
  // done during render, so a hidden menu is never painted against the new
  // state.
  const countdownKey = `${menuDisplayMode}|${submenuOpenCount}|${isMenuExpanded}|${isMenuMoved}`;
  const [prevCountdownKey, setPrevCountdownKey] = useState(countdownKey);
  if (countdownKey !== prevCountdownKey) {
    setPrevCountdownKey(countdownKey);
    setIsMenuHidden(false);
  }

  // Handle display mode changes
  useEffect(() => {
    if (menuDisplayMode === 'auto') {
      startCountdown();
    } else {
      stopCountdown();
    }
  }, [menuDisplayMode, startCountdown, stopCountdown]);

  // Handle hover and submenu state
  useEffect(() => {
    if (submenuOpenCount === 0 && !isMenuHovered) {
      startCountdown();
    } else {
      stopCountdown();
    }
  }, [submenuOpenCount, isMenuHovered, startCountdown, stopCountdown]);

  // Handle hover state
  const handleHovered = useCallback((hovered: boolean) => {
    setIsMenuHovered(hovered);
    if (hovered) {
      setIsMenuHidden(false);
    }
  }, []);

  const handleMoved = useCallback(() => {
    if (!isMenuMoved) {
      setIsMenuMoved(true);
    }
  }, [isMenuMoved]);

  return {
    isMenuExpanded,
    isMenuHidden,
    isMenuMoved,
    isInitialized,
    handleHovered,
    handleMoved,
    setIsMenuExpanded
  };
}
