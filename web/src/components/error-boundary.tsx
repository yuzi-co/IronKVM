import { ReactNode } from 'react';
import { Button, Tooltip } from 'antd';
import { TriangleAlertIcon } from 'lucide-react';
import { ErrorBoundary, type FallbackProps } from 'react-error-boundary';
import { useTranslation } from 'react-i18next';
import { isRouteErrorResponse, useRouteError } from 'react-router';

// The app-level boundary in main.tsx catches anything these do not, and what it
// does with it is replace the entire desktop with an error page. For a KVM that
// is the wrong trade almost every time: a settings popover that throws should
// not take away the keyboard, the mouse and the picture with it.
//
// So each of these keeps the failure inside the thing that failed, and each
// offers a way out. Retry re-renders the subtree, which is enough for a fault
// that has since cleared - a stream that reconnected, a device list that has
// arrived. Reload is the escape hatch for one that has not.

type BoundaryProps = {
  // Named so a crash says which part of the desktop failed, in the console and
  // to whoever is reading the screenshot afterwards.
  name: string;
  children: ReactNode;
};

function describe(error: unknown) {
  if (isRouteErrorResponse(error)) return `${error.status} ${error.statusText}`;
  if (error instanceof Error) return error.message;
  return error ? String(error) : '';
}

function logFailure(name: string) {
  return (error: unknown) => console.error(`[nanokvm] ${name} failed to render`, error);
}

function Actions({ retry, compact }: { retry: () => void; compact?: boolean }) {
  const { t } = useTranslation();

  return (
    <div className="flex items-center gap-2">
      <Button size="small" onClick={retry}>
        {t('error.retry')}
      </Button>
      {!compact && (
        <Button size="small" type="primary" danger onClick={() => window.location.reload()}>
          {t('error.refresh')}
        </Button>
      )}
    </div>
  );
}

const PanelFallback = ({ error, resetErrorBoundary }: FallbackProps) => {
  const { t } = useTranslation();
  const detail = describe(error);

  return (
    <div
      role="alert"
      className="flex h-full min-h-0 w-full min-w-0 flex-col items-center justify-center gap-3 p-4 text-center"
    >
      <TriangleAlertIcon className="h-6 w-6 shrink-0 text-amber-400" />
      <div className="text-sm text-neutral-300">{t('error.panel')}</div>
      {detail && <div className="max-w-full text-xs break-words text-neutral-500">{detail}</div>}
      <Actions retry={resetErrorBoundary} />
    </div>
  );
};

// The menubar is a fixed row of 30px squares, and it is the operator's only
// control surface. A button that throws keeps its square and becomes a retry
// button, so the bar neither collapses nor shifts every other item sideways
// while one of them is broken.
const MenuFallback = ({ resetErrorBoundary }: FallbackProps) => {
  const { t } = useTranslation();

  return (
    <Tooltip title={t('error.panel')} placement="bottom">
      <div
        role="alert"
        className="flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded text-amber-400 hover:bg-neutral-700/80 hover:text-amber-300"
        onClick={resetErrorBoundary}
      >
        <TriangleAlertIcon size={18} />
      </div>
    </Tooltip>
  );
};

// An overlay that fails must not paint a box over the picture it was
// decorating, so its fallback is a small badge pinned out of the way of the
// remote desktop.
const OverlayFallback = ({ resetErrorBoundary }: FallbackProps) => {
  const { t } = useTranslation();

  return (
    <div
      role="alert"
      className="fixed right-3 bottom-3 z-1100 flex items-center gap-2 rounded border border-amber-500/40 bg-neutral-900/90 px-3 py-2 text-xs text-amber-300 shadow-lg"
    >
      <TriangleAlertIcon size={14} className="shrink-0" />
      <span>{t('error.panel')}</span>
      <Actions retry={resetErrorBoundary} compact />
    </div>
  );
};

export const PanelBoundary = ({ name, children }: BoundaryProps) => (
  <ErrorBoundary FallbackComponent={PanelFallback} onError={logFailure(name)}>
    {children}
  </ErrorBoundary>
);

export const MenuBoundary = ({ name, children }: BoundaryProps) => (
  <ErrorBoundary FallbackComponent={MenuFallback} onError={logFailure(name)}>
    {children}
  </ErrorBoundary>
);

export const OverlayBoundary = ({ name, children }: BoundaryProps) => (
  <ErrorBoundary FallbackComponent={OverlayFallback} onError={logFailure(name)}>
    {children}
  </ErrorBoundary>
);

// react-router renders its own bare "Unexpected Application Error!" page for
// anything a route throws, and the app-level boundary in main.tsx never sees it
// because the router catches it first. A lazy chunk that fails to load is the
// case that matters: the device is reachable, the page is up, and the operator
// is looking at a white page with a stack trace and no way back.
export const RouteError = () => {
  const error = useRouteError();
  const { t } = useTranslation();
  const detail = describe(error);

  console.error('[nanokvm] route failed', error);

  return (
    <div
      role="alert"
      className="flex h-screen w-screen flex-col items-center justify-center space-y-4 bg-neutral-950 px-6 text-center"
    >
      <TriangleAlertIcon className="h-8 w-8 text-red-500" />
      <h2 className="text-lg font-semibold text-red-500">{t('error.title')}</h2>
      {detail && <div className="max-w-full text-xs break-words text-neutral-500">{detail}</div>}
      <Button type="primary" danger onClick={() => window.location.reload()}>
        {t('error.refresh')}
      </Button>
    </div>
  );
};
