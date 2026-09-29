import { Fragment, ReactNode, useEffect, useRef } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Divider } from 'antd';
import clsx from 'clsx';
import { useAtomValue, useSetAtom } from 'jotai';
import { GripVerticalIcon } from 'lucide-react';
import Draggable, { DraggableData, DraggableEvent } from 'react-draggable';
import { useMediaQuery } from 'react-responsive';

import { hasAudioAtom } from '@/jotai/audio.ts';
import { ocrSelectingAtom } from '@/jotai/ocr.ts';
import { inputRegionSelectingAtom } from '@/jotai/screen.ts';
import {
  keyboardLedStatusVisibleAtom,
  menuCloseSignalAtom,
  menuDisabledItemsAtom
} from '@/jotai/settings.ts';
import { useMenuBounds } from '@/hooks/useMenuBounds.ts';
import { useMenuVisibility } from '@/hooks/useMenuVisibility.ts';
import { MenuBoundary } from '@/components/error-boundary';

import { KeyboardLedStatus } from '../keyboard-led-status';
import { Alerts } from './alerts';
import { Fullscreen } from './fullscreen';
import { Keyboard } from './keyboard';
import { Media } from './media';
import { More } from './more.tsx';
import { Mouse } from './mouse';
import { Collapse, Expand } from './operations';
import { Picoclaw } from './picoclaw';
import { Power } from './power';
import { Screen } from './screen';
import { Script } from './script';
import { Settings } from './settings';
import { Speaker } from './speaker';
import { Terminal } from './terminal';
import { Text } from './text';
import { Tools } from './tools';
import { Wol } from './wol';

export const Menu = () => {
  const nodeRef = useRef<HTMLDivElement | null>(null);
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';

  const menuDisabledItems = useAtomValue(menuDisabledItemsAtom);
  const menuCloseSignal = useAtomValue(menuCloseSignalAtom);
  const isKeyboardLedStatusVisible = useAtomValue(keyboardLedStatusVisibleAtom);
  const hasAudio = useAtomValue(hasAudioAtom);
  const requestMenuClose = useSetAtom(menuCloseSignalAtom);

  // Below the sm breakpoint the full bar is wider than a phone. The entries
  // used most stay on the bar and the rest move into the overflow popover.
  const isNarrow = useMediaQuery({ maxWidth: 639 });

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

  // OCR and the input region selection collapse the bar so it is out of the
  // way of the drag. The bar comes back once the selection ends, if it was
  // open when the selection began; the collapse above alone left it folded.
  const isOcrSelecting = useAtomValue(ocrSelectingAtom);
  const isRegionSelecting = useAtomValue(inputRegionSelectingAtom);
  const isSelecting = isOcrSelecting || isRegionSelecting;
  const collapsedForSelection = useRef(false);
  useEffect(() => {
    if (isSelecting) {
      if (isMenuExpanded) collapsedForSelection.current = true;
      setIsMenuExpanded(false);
    } else if (collapsedForSelection.current) {
      collapsedForSelection.current = false;
      setIsMenuExpanded(true);
    }
    // Only the start and end of a selection matter here, not later toggles.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isSelecting]);

  function onDragStop(_e: DraggableEvent, data: DraggableData) {
    if (data.x === 0 && data.y === 0) return;
    handleMoved();
  }

  function isEnabled(item: string) {
    return !menuDisabledItems.includes(item);
  }

  const screen = (
    <MenuBoundary key="screen" name="screen">
      <Screen />
    </MenuBoundary>
  );
  const keyboard = (
    <MenuBoundary key="keyboard" name="keyboard">
      <Keyboard />
    </MenuBoundary>
  );
  const mouse = (
    <MenuBoundary key="mouse" name="mouse">
      <Mouse />
    </MenuBoundary>
  );
  const text = isEnabled('text') && (
    <MenuBoundary key="text" name="text">
      <Text />
    </MenuBoundary>
  );
  // hasAudio is set by the video path when sound arrives: the WebRTC audio
  // track, or the first audio frame on H.264 direct. A device without the USB
  // audio gadget sends neither, and the button would then unmute nothing.
  const speaker = isEnabled('speaker') && hasAudio && (
    <MenuBoundary key="speaker" name="speaker">
      <Speaker />
    </MenuBoundary>
  );
  const media = isAdmin && isEnabled('media') && (
    <MenuBoundary key="media" name="media">
      <Media />
    </MenuBoundary>
  );
  const terminal = isAdmin && isEnabled('terminal') && (
    <MenuBoundary key="terminal" name="terminal">
      <Terminal />
    </MenuBoundary>
  );

  // The Tools menu shows the entries this account may use and has not hidden,
  // and leaves the bar when none is left.
  const script = isAdmin && isEnabled('script') && (
    <MenuBoundary name="script">
      <Script />
    </MenuBoundary>
  );
  const wol = isEnabled('wol') && (
    <MenuBoundary name="wol">
      <Wol />
    </MenuBoundary>
  );
  const picoclaw = isAdmin && isEnabled('picoclaw') && (
    <MenuBoundary name="picoclaw">
      <Picoclaw />
    </MenuBoundary>
  );
  const tools = isEnabled('tools') && (script || wol || picoclaw) && (
    <MenuBoundary key="tools" name="tools">
      <Tools>
        {script}
        {wol}
        {picoclaw}
      </Tools>
    </MenuBoundary>
  );

  const power = isEnabled('power') && (
    <MenuBoundary key="power" name="power">
      <Power />
    </MenuBoundary>
  );
  const settings = (
    <MenuBoundary key="settings" name="settings">
      <Settings />
    </MenuBoundary>
  );
  const fullscreen = isEnabled('fullscreen') && (
    <MenuBoundary key="fullscreen" name="fullscreen">
      <Fullscreen />
    </MenuBoundary>
  );
  const collapse = (toggleMenu: (expanded: boolean) => void) =>
    isEnabled('collapse') && (
      <MenuBoundary key="collapse" name="collapse">
        <Collapse toggleMenu={toggleMenu} />
      </MenuBoundary>
    );

  // The wide bar in groups: the picture, input, media and tools, the host's
  // power, and the bar itself. A separator stands between two groups that
  // both have something to show.
  const groups: ReactNode[][] = [
    [screen, speaker],
    [keyboard, mouse, text],
    [media, terminal, tools],
    [power],
    [settings, fullscreen, collapse(setIsMenuExpanded)]
  ]
    .map((group) => group.filter(Boolean))
    .filter((group) => group.length > 0);

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
          'fixed top-[10px] left-1/2 z-1000 transition-opacity duration-300',
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
            {isMenuExpanded && isKeyboardLedStatusVisible && !isNarrow && (
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
            {/* Shown only while something is wrong. */}
            <MenuBoundary name="alerts">
              <Alerts />
            </MenuBoundary>
            <strong>
              <div className="flex h-[30px] cursor-move items-center justify-center pl-1 text-neutral-500 select-none">
                <GripVerticalIcon size={18} />
              </div>
            </strong>
            <Divider type="vertical" />

            {isNarrow ? (
              <>
                {keyboard}
                {mouse}
                {media}
                {power}
                <Divider type="vertical" />
                {settings}
                <More>
                  {screen}
                  {speaker}
                  {terminal}
                  {text}
                  {tools}
                  {fullscreen}
                  {collapse(() => requestMenuClose((signal) => signal + 1))}
                </More>
              </>
            ) : (
              groups.map((group, index) => (
                <Fragment key={index}>
                  {index > 0 && <Divider type="vertical" />}
                  {group}
                </Fragment>
              ))
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
