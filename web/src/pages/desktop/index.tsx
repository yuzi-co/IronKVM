import { lazy, Suspense, useEffect, useState } from 'react';
import { Splitter } from 'antd';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from 'react-responsive';

import { getInputRegion, getScreen, setControlRegionMode } from '@/api/vm.ts';
import { ControlRegionConfig, InputRegion, ScreenSettings } from '@/types';
import * as storage from '@/lib/localstorage.ts';
import { client } from '@/lib/websocket.ts';
import { isKeyboardOpenAtom } from '@/jotai/keyboard.ts';
import { picoclawChatOpenAtom } from '@/jotai/picoclaw.ts';
import {
  captureStatusAtom,
  controlRegionModeAtom,
  inputRegionAtom,
  manualInputRegionAtom,
  manualRegionsAtom,
  resolutionAtom,
  screenSettingsAtom,
  selectedManualRegionAtom,
  selectedOriginalResolutionAtom,
  videoModeAtom
} from '@/jotai/screen.ts';
import { OverlayBoundary, PanelBoundary } from '@/components/error-boundary';
import { Head } from '@/components/head.tsx';

import { CaptureStatusOverlay, useCaptureStatus } from './capture-status';
import { AbsoluteMouseWarning, InputDisconnectedWarning } from './hid-status';
import { IonCheckingIndicator, IonCriticalGate, IonWarningBadge, useIonStatus } from './ion-status';
import { Keyboard } from './keyboard';
import { Menu } from './menu';
import { Mouse } from './mouse';
import { H264ModeNotification, Notification } from './notification.tsx';
import { Ocr } from './ocr';
import { Paste } from './paste';
import { ActionOverlay } from './picoclaw/action-overlay.tsx';
import { Screen } from './screen';
import { AutoRegion } from './screen/auto-region.tsx';
import {
  getMediaSize,
  isInputRegionCompatible,
  isMediaReady,
  isValidInputRegion
} from './screen/geometry.ts';
import { InputRegionOverlay } from './screen/input-region-overlay.tsx';
import { ManualRegion } from './screen/manual-region.tsx';
import { usePauseWhenHidden } from './screen/use-pause-when-hidden.ts';
import { ViewOnlyBadge } from './view-only-badge.tsx';

// H.264 direct is the default where the browser can decode both its video and
// its audio. Measured on 2026-09-23 at 1080p30, it held the board at 17% busy
// against 28.5% for WebRTC, because WebRTC seals and sends every 1200 bytes as
// its own packet on a core with no AES instructions. Direct carries the host's
// audio too, but only through WebCodecs' AudioDecoder, so a browser without it
// keeps WebRTC and does not lose the sound.
function getDefaultVideoMode() {
  if (window.VideoDecoder && window.AudioDecoder) {
    return 'direct';
  }

  return window.RTCPeerConnection ? 'h264' : 'mjpeg';
}

function getVideoMode() {
  const defaultVideoMode = getDefaultVideoMode();

  const cookieVideoMode = storage.getVideoMode();
  if (!cookieVideoMode || (cookieVideoMode === 'direct' && !window.VideoDecoder)) {
    return defaultVideoMode;
  }

  return ['direct', 'h264', 'mjpeg'].includes(cookieVideoMode) ? cookieVideoMode : defaultVideoMode;
}

// The PicoClaw chat, with its markdown renderer, and the virtual keyboard are
// opened in few sessions, so each loads the first time it is asked for.
const PicoclawSidebar = lazy(() => import('./picoclaw').then((m) => ({ default: m.Sidebar })));
const VirtualKeyboard = lazy(() =>
  import('./virtual-keyboard').then((m) => ({ default: m.VirtualKeyboard }))
);

const PicoclawLoading = () => (
  <div className="flex h-full w-full items-center justify-center text-neutral-500">
    <LoaderCircleIcon className="animate-spin" size={18} />
  </div>
);

// Mounted from the first open on, and kept mounted after, so that closing
// the keyboard still plays the drawer's exit and its layout choice survives.
const LazyVirtualKeyboard = () => {
  const isKeyboardOpen = useAtomValue(isKeyboardOpenAtom);
  const [wanted, setWanted] = useState(isKeyboardOpen);
  if (isKeyboardOpen && !wanted) setWanted(true);

  if (!wanted) return null;
  return (
    <Suspense fallback={null}>
      <VirtualKeyboard />
    </Suspense>
  );
};

export const Desktop = () => {
  const { t } = useTranslation();
  const isBigScreen = useMediaQuery({ minWidth: 850 });
  const [activeVideoMode] = useState(getVideoMode);
  const [picoclawSidebarWidth, setPicoclawSidebarWidth] = useState(420);
  const captureStatus = useCaptureStatus(activeVideoMode);
  const ion = useIonStatus();
  const isStreamPaused = usePauseWhenHidden();
  const setCaptureStatus = useSetAtom(captureStatusAtom);

  // The toolbar's Screen dot and alert icon read the same status.
  useEffect(() => {
    setCaptureStatus(captureStatus);
  }, [captureStatus, setCaptureStatus]);

  const [videoMode, setVideoMode] = useAtom(videoModeAtom);
  const [resolution, setResolution] = useAtom(resolutionAtom);
  const [inputRegion, setInputRegion] = useAtom(inputRegionAtom);
  const [controlRegionMode, setControlRegionModeState] = useAtom(controlRegionModeAtom);
  const setManualInputRegion = useSetAtom(manualInputRegionAtom);
  const setManualRegions = useSetAtom(manualRegionsAtom);
  const setSelectedManualRegion = useSetAtom(selectedManualRegionAtom);
  const setSelectedOriginalResolution = useSetAtom(selectedOriginalResolutionAtom);
  const setScreenSettings = useSetAtom(screenSettingsAtom);
  const isPicoclawChatOpen = useAtomValue(picoclawChatOpenAtom);

  useEffect(() => {
    client.connect();

    setVideoMode(activeVideoMode);

    // The capture resolution belongs to the board, not to this browser. It
    // starts at auto and is corrected as soon as the server answers, because
    // absolute mouse positioning scales by it and a wrong value puts the
    // pointer in the wrong place.
    setResolution({ width: 0, height: 0 });
    setInputRegion(null);
    setManualInputRegion(null);
    setManualRegions([]);
    setSelectedManualRegion('');
    setSelectedOriginalResolution('');
    setControlRegionModeState('off');

    getScreen()
      .then((rsp) => {
        if (rsp.code !== 0) return;

        const settings = rsp.data as ScreenSettings | null;
        if (!settings) return;

        setResolution({ width: settings.width, height: settings.height });
        setScreenSettings(settings);
      })
      .catch(() => {
        // Auto is the safe answer: the capture pipeline picks the source
        // resolution and the mouse scales to what the picture reports.
      });

    getInputRegion()
      .then((rsp) => {
        const config = rsp.data as ControlRegionConfig | null;
        const mode = config?.mode || 'off';
        const manualRegion = isValidInputRegion(config as InputRegion)
          ? (config as InputRegion)
          : null;
        const selectedResolution = config?.selectedResolution || '';
        const regions = config?.regions || [];
        const selectedRegion = config?.selectedRegion || '';
        const selectedManualRegion = regions.find(
          (region) => `${region.width}x${region.height}` === selectedRegion
        );
        setManualInputRegion(manualRegion);
        setManualRegions(regions);
        setSelectedManualRegion(selectedRegion);
        setSelectedOriginalResolution(selectedResolution);
        setInputRegion(mode === 'manual' && selectedRegion ? selectedManualRegion || null : null);
        setControlRegionModeState(mode);
      })
      .catch(() => {
        setControlRegionModeState('off');
        setInputRegion(null);
        setManualInputRegion(null);
        setManualRegions([]);
        setSelectedManualRegion('');
        setSelectedOriginalResolution('');
      });

    return () => {
      client.close();
    };
  }, [
    activeVideoMode,
    setControlRegionModeState,
    setInputRegion,
    setManualInputRegion,
    setManualRegions,
    setResolution,
    setSelectedManualRegion,
    setScreenSettings,
    setSelectedOriginalResolution,
    setVideoMode
  ]);

  useEffect(() => {
    if (controlRegionMode !== 'manual' || !inputRegion) {
      return;
    }

    const screen = document.getElementById('screen');
    if (!screen) {
      return;
    }
    const target = screen;
    const region = inputRegion;

    let cleared = false;
    let validationTimer: ReturnType<typeof setTimeout> | null = null;
    function validateMediaSize() {
      if (cleared) {
        return;
      }

      if (validationTimer !== null) {
        clearTimeout(validationTimer);
      }
      validationTimer = setTimeout(() => {
        const mediaSize = getMediaSize(target, resolution);
        if (!mediaSize || !isMediaReady(target) || isInputRegionCompatible(region, mediaSize)) {
          return;
        }

        cleared = true;
        setControlRegionMode('off')
          .then((rsp) => {
            if (rsp.code === 0) {
              setControlRegionModeState('off');
              setInputRegion(null);
            }
          })
          .catch(() => undefined);
      }, 300);
    }

    validateMediaSize();
    const observer = new MutationObserver(validateMediaSize);
    observer.observe(target, {
      attributes: true,
      attributeFilter: ['data-media-width', 'data-media-height']
    });
    target.addEventListener('load', validateMediaSize);
    target.addEventListener('loadedmetadata', validateMediaSize);
    target.addEventListener('canplay', validateMediaSize);
    target.addEventListener('resize', validateMediaSize);

    return () => {
      observer.disconnect();
      if (validationTimer !== null) {
        clearTimeout(validationTimer);
      }
      target.removeEventListener('load', validateMediaSize);
      target.removeEventListener('loadedmetadata', validateMediaSize);
      target.removeEventListener('canplay', validateMediaSize);
      target.removeEventListener('resize', validateMediaSize);
    };
  }, [
    controlRegionMode,
    inputRegion,
    resolution,
    setControlRegionModeState,
    setInputRegion,
    videoMode
  ]);

  function handleSplitterResize(sizes: number[]) {
    const nextSidebarWidth = sizes[1];
    if (typeof nextSidebarWidth === 'number' && nextSidebarWidth > 0) {
      setPicoclawSidebarWidth(nextSidebarWidth);
    }
  }

  return (
    <div className="h-screen w-screen overflow-hidden bg-neutral-950">
      <Head title={t('head.desktop')} />

      <OverlayBoundary name="notifications">
        {isBigScreen && <Notification />}
        <H264ModeNotification />
        <AbsoluteMouseWarning />
        <InputDisconnectedWarning />
        <ViewOnlyBadge />
      </OverlayBoundary>

      {videoMode && resolution && (
        <div className="relative flex h-full min-h-0 w-full min-w-0">
          <Menu />
          <div className="h-full min-h-0 w-full min-w-0">
            {/* With the chat closed the bar sits on the screen's right edge,
                and its disabled dragger, 6 px wide around it, took the
                pointer from the screen's last 3 px. */}
            <Splitter
              className={
                isBigScreen && isPicoclawChatOpen
                  ? 'h-full w-full'
                  : 'h-full w-full [&>.ant-splitter-bar]:hidden'
              }
              style={{ height: '100%', width: '100%' }}
              onResize={handleSplitterResize}
            >
              <Splitter.Panel min="45%">
                <div className="relative h-full min-h-0 w-full min-w-0 overflow-hidden bg-black">
                  {ion.holdStream ? (
                    ion.loading ? (
                      <IonCheckingIndicator />
                    ) : (
                      ion.status?.verdict === 'critical' && (
                        <IonCriticalGate onContinue={ion.continueAnyway} />
                      )
                    )
                  ) : (
                    <>
                      <PanelBoundary name="screen">{!isStreamPaused && <Screen />}</PanelBoundary>
                      <OverlayBoundary name="capture-status">
                        <CaptureStatusOverlay status={captureStatus} />
                      </OverlayBoundary>
                      <OverlayBoundary name="ion-status">
                        <IonWarningBadge status={ion.status} />
                      </OverlayBoundary>
                    </>
                  )}
                </div>
              </Splitter.Panel>
              <Splitter.Panel
                size={isBigScreen && isPicoclawChatOpen ? picoclawSidebarWidth : 0}
                min={isBigScreen && isPicoclawChatOpen ? 340 : 0}
                max="45%"
                resizable={isBigScreen && isPicoclawChatOpen}
              >
                {isBigScreen && isPicoclawChatOpen ? (
                  <PanelBoundary name="picoclaw-sidebar">
                    <Suspense fallback={<PicoclawLoading />}>
                      <PicoclawSidebar />
                    </Suspense>
                  </PanelBoundary>
                ) : null}
              </Splitter.Panel>
            </Splitter>
          </div>
          <OverlayBoundary name="regions">
            <ActionOverlay />
            <AutoRegion />
            <ManualRegion />
            <InputRegionOverlay />
          </OverlayBoundary>
          <OverlayBoundary name="mouse">
            <Mouse />
          </OverlayBoundary>
          <OverlayBoundary name="keyboard">
            <Keyboard />
          </OverlayBoundary>
        </div>
      )}

      {!isBigScreen && isPicoclawChatOpen ? (
        <div className="fixed inset-x-0 top-14 bottom-0 z-980 overflow-hidden bg-[#0d0d0f] shadow-2xl">
          <PanelBoundary name="picoclaw-sidebar">
            <Suspense fallback={<PicoclawLoading />}>
              <PicoclawSidebar />
            </Suspense>
          </PanelBoundary>
        </div>
      ) : null}

      <OverlayBoundary name="virtual-keyboard">
        <LazyVirtualKeyboard />
      </OverlayBoundary>

      <OverlayBoundary name="paste">
        <Paste />
      </OverlayBoundary>

      <OverlayBoundary name="ocr">
        <Ocr />
      </OverlayBoundary>
    </div>
  );
};
