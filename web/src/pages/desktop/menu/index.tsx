import { useEffect, useRef } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Divider } from 'antd';
import clsx from 'clsx';
import { useAtomValue } from 'jotai';
import { GripVerticalIcon } from 'lucide-react';
import Draggable, { DraggableData, DraggableEvent } from 'react-draggable';

import { hasAudioAtom } from '@/jotai/audio.ts';
import {
  keyboardLedStatusVisibleAtom,
  menuCloseSignalAtom,
  menuDisabledItemsAtom
} from '@/jotai/settings.ts';
import { useMenuBounds } from '@/hooks/useMenuBounds.ts';
import { useMenuVisibility } from '@/hooks/useMenuVisibility.ts';
import { MenuBoundary } from '@/components/error-boundary';

import { KeyboardLedStatus } from '../keyboard-led-status';
import { DownloadImage } from './download.tsx';
import { Fullscreen } from './fullscreen';
import { Image } from './image';
import { Keyboard } from './keyboard';
import { Mouse } from './mouse';
import { Collapse, Expand } from './operations';
import { Picoclaw } from './picoclaw';
import { Power } from './power';
import { Screen } from './screen';
import { Script } from './script';
import { Settings } from './settings';
import { Speaker } from './speaker';
import { Terminal } from './terminal';
import { Wol } from './wol';

export const Menu = () => {
  const nodeRef = useRef<HTMLDivElement | null>(null);
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';

  const menuDisabledItems = useAtomValue(menuDisabledItemsAtom);
  const menuCloseSignal = useAtomValue(menuCloseSignalAtom);
  const isKeyboardLedStatusVisible = useAtomValue(keyboardLedStatusVisibleAtom);
  const hasAudio = useAtomValue(hasAudioAtom);

  const {
    isInitialized,
    isMenuExpanded,
    isMenuHidden,
    handleHovered,
    handleMoved,
    setIsMenuExpanded
  } = useMenuVisibility();

  const menuBounds = useMenuBounds(nodeRef, isMenuExpanded);

  useEffect(() => {
    if (menuCloseSignal > 0) {
      setIsMenuExpanded(false);
    }
  }, [menuCloseSignal, setIsMenuExpanded]);

  function onDragStop(_e: DraggableEvent, data: DraggableData) {
    if (data.x === 0 && data.y === 0) return;
    handleMoved();
  }

  function isEnabled(item: string) {
    return !menuDisabledItems.includes(item);
  }

  return (
    <Draggable
      nodeRef={nodeRef}
      bounds={menuBounds}
      handle="strong"
      positionOffset={{ x: '-50%', y: '0%' }}
      onStop={onDragStop}
    >
      <div
        ref={nodeRef}
        className={clsx(
          'fixed top-[10px] left-1/2 z-1000 -translate-x-1/2 transition-opacity duration-300',
          isInitialized ? 'opacity-100' : 'opacity-0'
        )}
        onMouseEnter={() => handleHovered(true)}
        onMouseLeave={() => handleHovered(false)}
        onBlur={() => handleHovered(false)}
      >
        {/* Trigger area for auto-show when hidden */}
        {isMenuExpanded && (
          <div className="absolute top-[-10px] right-0 left-0 h-[46px] w-full bg-transparent" />
        )}

        {/* Menubar */}
        <div className="sticky top-[10px] flex w-full justify-center">
          <div
            className={clsx(
              'relative h-[36px] items-center rounded bg-neutral-800/80 pr-2 pl-1 transition-all duration-300',
              isMenuExpanded ? 'flex' : 'hidden',
              isMenuHidden ? 'translate-y-[-110%] opacity-80' : 'translate-y-0 opacity-100'
            )}
          >
            {isMenuExpanded && isKeyboardLedStatusVisible && (
              <div
                className={clsx(
                  'absolute inset-y-0 right-full mr-1 transition-all duration-300',
                  isMenuHidden ? 'pointer-events-none opacity-0' : 'opacity-100'
                )}
              >
                <MenuBoundary name="keyboard-led-status">
                  <KeyboardLedStatus />
                </MenuBoundary>
              </div>
            )}
            <strong>
              <div className="flex h-[30px] cursor-move items-center justify-center pl-1 text-neutral-500 select-none">
                <GripVerticalIcon size={18} />
              </div>
            </strong>
            <Divider type="vertical" />

            <MenuBoundary name="screen">
              <Screen />
            </MenuBoundary>
            <MenuBoundary name="keyboard">
              <Keyboard />
            </MenuBoundary>
            <MenuBoundary name="mouse">
              <Mouse />
            </MenuBoundary>
            {/* hasAudio comes from the peer connection's ontrack event. A
                device without the USB audio gadget offers no audio track, and
                the button would then unmute an <audio> with no stream in it. */}
            {isEnabled('speaker') && hasAudio && (
              <MenuBoundary name="speaker">
                <Speaker />
              </MenuBoundary>
            )}
            <Divider type="vertical" />

            {isAdmin && isEnabled('image') && (
              <MenuBoundary name="image">
                <Image />
              </MenuBoundary>
            )}
            {isAdmin && isEnabled('download') && (
              <MenuBoundary name="download">
                <DownloadImage />
              </MenuBoundary>
            )}
            {isAdmin && isEnabled('terminal') && (
              <MenuBoundary name="terminal">
                <Terminal />
              </MenuBoundary>
            )}
            {isAdmin && isEnabled('script') && (
              <MenuBoundary name="script">
                <Script />
              </MenuBoundary>
            )}
            {isEnabled('wol') && (
              <MenuBoundary name="wol">
                <Wol />
              </MenuBoundary>
            )}

            {(isEnabled('wol') ||
              (isAdmin && ['image', 'download', 'script', 'terminal'].some(isEnabled))) && (
              <Divider type="vertical" />
            )}

            {isAdmin && isEnabled('picoclaw') && (
              <>
                <MenuBoundary name="picoclaw">
                  <Picoclaw />
                </MenuBoundary>
                <Divider type="vertical" />
              </>
            )}

            {isEnabled('power') && (
              <>
                <MenuBoundary name="power">
                  <Power />
                </MenuBoundary>
                <Divider type="vertical" />
              </>
            )}

            <MenuBoundary name="settings">
              <Settings />
            </MenuBoundary>
            {isEnabled('fullscreen') && (
              <MenuBoundary name="fullscreen">
                <Fullscreen />
              </MenuBoundary>
            )}
            {isEnabled('collapse') && (
              <MenuBoundary name="collapse">
                <Collapse toggleMenu={setIsMenuExpanded} />
              </MenuBoundary>
            )}
          </div>
        </div>

        {/* Menubar expand button */}
        {!isMenuExpanded && (
          <MenuBoundary name="expand">
            <Expand toggleMenu={setIsMenuExpanded} />
          </MenuBoundary>
        )}
      </div>
    </Draggable>
  );
};
