import { useEffect, useRef, useState, useSyncExternalStore } from 'react';
import type { ReactNode } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Badge, Input, Modal, Tooltip } from 'antd';
import clsx from 'clsx';
import { useAtom, useSetAtom } from 'jotai';
import {
  BadgeInfoIcon,
  BotIcon,
  CircleArrowUpIcon,
  DiscIcon,
  GaugeIcon,
  HeartPulseIcon,
  KeyRoundIcon,
  LockIcon,
  LockKeyholeIcon,
  MonitorDownIcon,
  NetworkIcon,
  PaletteIcon,
  PowerIcon,
  ScreenShareIcon,
  SearchIcon,
  ServerCogIcon,
  SettingsIcon,
  ShieldIcon,
  SmartphoneIcon,
  TerminalIcon,
  UserRoundIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import semver from 'semver';

import * as api from '@/api/application.ts';
import * as ls from '@/lib/localstorage.ts';
import { keyboardLockAtom } from '@/jotai/keyboard.ts';
import { settingsOpenRequestAtom, submenuOpenCountAtom } from '@/jotai/settings.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';
import { ScrollArea } from '@/components/ui/scroll-area';

import { About } from './about';
import { Account } from './account';
import { APIKeys } from './api-keys';
import { Device } from './device';
import { Ipmi } from './ipmi';
import { MCP } from './mcp';
import { VirtualMedia } from './media';
import { SettingsNav } from './nav-context.ts';
import {
  browserStorage,
  filterTabs,
  groupTabs,
  initialTab,
  LAST_TAB_KEY,
  readStored,
  writeStored
} from './nav.ts';
import type { Group } from './nav.ts';
import { Netboot } from './netboot';
import { Network } from './network';
import { Performance } from './performance';
import { Preferences } from './preferences';
import { Redfish } from './redfish';
import { Ssh } from './ssh';
import { Tls } from './tls';
import { Update } from './update';
import { Vnc } from './vnc';
import { VpnTab } from './vpn/tab.tsx';
import { Watchdog } from './watchdog';

type Tab = {
  id: string;
  group: Group;
  icon: ReactNode;
  // A name that needs no translation; otherwise settings.<id>.title.
  label?: string;
  component: ReactNode;
};

// Tailwind's sm breakpoint, which is where the sidebar drops its labels.
const narrowQuery = '(max-width: 639.98px)';

function subscribeNarrow(onChange: () => void) {
  const query = window.matchMedia(narrowQuery);
  query.addEventListener('change', onChange);
  return () => query.removeEventListener('change', onChange);
}

function useIsNarrow() {
  return useSyncExternalStore(subscribeNarrow, () => window.matchMedia(narrowQuery).matches);
}

export const Settings = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';
  // Below sm the sidebar shows icons only, so each item needs a tooltip.
  const isNarrow = useIsNarrow();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [currentTab, setCurrentTab] = useState('about');
  const [query, setQuery] = useState('');
  const scrollViewportRef = useRef<HTMLDivElement>(null);

  const [isUpdateAvailable, setIsUpdateAvailable] = useState(false);
  const setKeyboardLock = useSetAtom(keyboardLockAtom);
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);

  const icon16 = { size: 16 };
  // The sidebar orders tabs by group, and within a group by this list. About
  // comes first so that a first visit opens on it, though the sidebar shows
  // it last, in the footer.
  const tabs: Tab[] = [
    { id: 'about', group: 'footer', icon: <BadgeInfoIcon {...icon16} />, component: <About /> },
    {
      id: 'account',
      group: 'access',
      icon: <UserRoundIcon {...icon16} />,
      component: <Account />
    },
    { id: 'apiKeys', group: 'access', icon: <KeyRoundIcon {...icon16} />, component: <APIKeys /> },
    {
      id: 'preferences',
      group: 'browser',
      icon: <PaletteIcon {...icon16} />,
      component: <Preferences />
    },
    ...(isAdmin
      ? ([
          {
            id: 'device',
            group: 'system',
            icon: <SmartphoneIcon {...icon16} />,
            component: <Device />
          },
          {
            id: 'performance',
            group: 'system',
            icon: <GaugeIcon {...icon16} />,
            component: <Performance />
          },
          {
            id: 'watchdog',
            group: 'system',
            icon: <HeartPulseIcon {...icon16} />,
            component: <Watchdog />
          },
          {
            id: 'update',
            group: 'system',
            icon: <CircleArrowUpIcon {...icon16} />,
            component: <Update setIsLocked={setIsLocked} />
          },
          {
            id: 'network',
            group: 'network',
            icon: <NetworkIcon {...icon16} />,
            component: <Network />
          },
          {
            id: 'vpn',
            group: 'network',
            icon: <ShieldIcon {...icon16} />,
            label: 'VPN',
            component: <VpnTab isLocked={isLocked} setIsLocked={setIsLocked} />
          },
          {
            id: 'ssh',
            group: 'access',
            icon: <TerminalIcon {...icon16} />,
            label: 'SSH',
            component: <Ssh />
          },
          { id: 'vnc', group: 'access', icon: <ScreenShareIcon {...icon16} />, component: <Vnc /> },
          {
            id: 'tls',
            group: 'access',
            icon: <LockKeyholeIcon {...icon16} />,
            label: 'TLS',
            component: <Tls setIsLocked={setIsLocked} />
          },
          {
            id: 'ipmi',
            group: 'integrations',
            icon: <PowerIcon {...icon16} />,
            component: <Ipmi />
          },
          {
            id: 'redfish',
            group: 'integrations',
            icon: <ServerCogIcon {...icon16} />,
            component: <Redfish />
          },
          { id: 'mcp', group: 'integrations', icon: <BotIcon {...icon16} />, component: <MCP /> },
          {
            id: 'netboot',
            group: 'boot',
            icon: <MonitorDownIcon {...icon16} />,
            component: <Netboot setIsLocked={setIsLocked} />
          },
          {
            id: 'media',
            group: 'boot',
            icon: <DiscIcon {...icon16} />,
            component: <VirtualMedia setIsLocked={setIsLocked} />
          }
        ] satisfies Tab[])
      : [])
  ];
  const groups = groupTabs(tabs);
  const footer = tabs.filter((tab) => tab.group === 'footer');
  const labelOf = (tab: Tab) => tab.label ?? t(`settings.${tab.id}.title`);
  // While searching, the sidebar is one flat list of the matches, in sidebar
  // order.
  const matches = query.trim()
    ? filterTabs(
        [...groups.flatMap((entry) => entry.tabs), ...footer].map((tab) => ({
          ...tab,
          label: labelOf(tab)
        })),
        query
      )
    : null;

  useEffect(() => {
    if (!isAdmin) return;
    const skip = ls.getSkipUpdate();
    if (skip) return;

    api.getVersion().then((rsp: any) => {
      if (rsp.code !== 0) {
        return;
      }
      if (!rsp.data?.current || !rsp.data?.latest) {
        return;
      }

      if (semver.gt(rsp.data.latest, rsp.data.current)) {
        setIsUpdateAvailable(true);
      }
    });
  }, [isAdmin]);

  useEffect(() => {
    scrollViewportRef.current?.scrollTo({ top: 0, left: 0 });
  }, [currentTab]);

  // showTab switches the page. Opening the update page counts as having seen
  // the update, which clears the badge.
  function showTab(tab: string) {
    setCurrentTab(tab);

    if (isUpdateAvailable && tab === 'update') {
      setIsUpdateAvailable(false);
      ls.setSkipUpdate(true);
    }
  }

  function changeTab(tab: string) {
    if (isLocked) {
      return;
    }

    showTab(tab);
    writeStored(browserStorage(), LAST_TAB_KEY, tab);
  }

  // openModal opens on the tab asked for, when this account has it, and
  // otherwise on the one initialTab picks.
  function openModal(requested?: string) {
    const ids = tabs.map((tab) => tab.id);
    setQuery('');
    if (requested && ids.includes(requested)) {
      changeTab(requested);
    } else {
      showTab(initialTab(ids, readStored(browserStorage(), LAST_TAB_KEY), isUpdateAvailable));
    }

    setIsModalOpen(true);
    setKeyboardLock({ source: 'settings-modal', locked: true });
    setSubmenuOpenCount((count) => count + 1);
  }

  // Another part of the UI asked for a settings page.
  const [openRequest, setOpenRequest] = useAtom(settingsOpenRequestAtom);
  const openModalStable = useStableCallback(openModal);
  useEffect(() => {
    if (!openRequest) return;
    setOpenRequest(null);
    if (!isModalOpen) openModalStable(openRequest);
  }, [openRequest, setOpenRequest, isModalOpen, openModalStable]);

  function closeModal() {
    if (isLocked) {
      return;
    }

    setKeyboardLock({ source: 'settings-modal', locked: false });
    setIsModalOpen(false);
    setSubmenuOpenCount((count) => Math.max(0, count - 1));
  }

  function renderItem(tab: Tab) {
    const label = labelOf(tab);
    const isCurrent = currentTab === tab.id;
    const isDisabled = isLocked && !isCurrent;

    let tip: string | undefined;
    if (isDisabled) {
      tip = t('settings.nav.locked');
    } else if (isNarrow) {
      tip = label;
    }

    const text = <span className="hidden truncate text-sm sm:block">{label}</span>;

    return (
      <Tooltip key={tab.id} title={tip} placement="right" mouseEnterDelay={0.3}>
        <button
          type="button"
          aria-label={label}
          aria-current={isCurrent ? 'page' : undefined}
          aria-disabled={isDisabled || undefined}
          className={clsx(
            'flex w-full shrink-0 items-center space-x-2 rounded-lg p-2 text-left select-none sm:px-3',
            isCurrent && 'bg-neutral-700/50',
            !isCurrent && !isDisabled && 'cursor-pointer hover:bg-neutral-700/50',
            isDisabled && 'cursor-not-allowed opacity-40'
          )}
          onClick={() => changeTab(tab.id)}
        >
          <div className="h-[16px] w-[16px] shrink-0">{tab.icon}</div>

          {isUpdateAvailable && tab.id === 'update' ? (
            <Badge dot color="blue" offset={[6, 3]}>
              {text}
            </Badge>
          ) : (
            text
          )}
        </button>
      </Tooltip>
    );
  }

  return (
    <>
      <Tooltip title={t('settings.title')} placement="bottom" mouseEnterDelay={0.6}>
        <button
          type="button"
          aria-label={t('settings.title')}
          className="flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 hover:bg-neutral-700/80"
          onClick={() => openModal()}
        >
          <Badge dot={isUpdateAvailable} color="blue" offset={[0, 2]}>
            <div className="pt-[3px] text-neutral-300 hover:text-white">
              <SettingsIcon size={18} />
            </div>
          </Badge>
        </button>
      </Tooltip>

      <Modal
        open={isModalOpen}
        width={'80%'}
        centered={true}
        footer={null}
        destroyOnHidden={true}
        closable={!isLocked}
        maskClosable={!isLocked}
        keyboard={!isLocked}
        onCancel={closeModal}
        style={{ maxWidth: '1080px' }}
        styles={{ container: { padding: 0 } }}
      >
        <div className="flex h-[80vh] max-h-[700px] rounded-lg outline outline-1 outline-neutral-700">
          <nav
            aria-label={t('settings.title')}
            className="box-border flex h-full max-w-[260px] shrink-0 flex-col overflow-y-auto rounded-l-lg bg-neutral-800/90 px-1 pb-4 sm:w-1/5 md:w-1/4 md:px-2"
          >
            <div className="hidden shrink-0 px-3 pt-10 text-xl sm:block">{t('settings.title')}</div>
            <div className="h-10 shrink-0 sm:h-2" />

            {isLocked && (
              <div className="mx-1 mb-2 hidden shrink-0 items-start space-x-2 rounded-md bg-amber-500/10 px-2 py-1.5 text-xs text-amber-400 sm:flex">
                <LockIcon size={12} className="mt-0.5 shrink-0" />
                <span>{t('settings.nav.locked')}</span>
              </div>
            )}

            {/* Below sm the sidebar is icons only: too narrow for a field. */}
            <div className="mx-1 hidden shrink-0 sm:block">
              <Input
                size="small"
                allowClear
                value={query}
                placeholder={t('settings.nav.search')}
                aria-label={t('settings.nav.search')}
                prefix={<SearchIcon size={14} className="text-neutral-500" />}
                onChange={(e) => setQuery(e.target.value)}
                onPressEnter={() => matches?.[0] && changeTab(matches[0].id)}
              />
            </div>

            {matches ? (
              <div className="flex flex-col space-y-0.5 pt-3">
                {matches.length > 0 ? (
                  matches.map(renderItem)
                ) : (
                  <div className="px-3 text-xs text-neutral-500">{t('settings.nav.noMatch')}</div>
                )}
              </div>
            ) : (
              <>
                {groups.map(({ group, tabs: groupItems }, index) => (
                  <div key={group} role="group" aria-label={t(`settings.nav.${group}`)}>
                    {/* Icon-only below sm: a rule stands in for the heading. */}
                    {index > 0 && (
                      <div className="mx-2 my-2 border-t border-neutral-700 sm:hidden" />
                    )}
                    <div className="hidden px-3 pt-4 pb-1 text-xs font-medium tracking-wide text-neutral-500 uppercase sm:block">
                      {t(`settings.nav.${group}`)}
                    </div>
                    <div className="flex flex-col space-y-0.5">{groupItems.map(renderItem)}</div>
                  </div>
                ))}

                {/* About is reference rather than a setting, so it sits apart
                    at the foot of the sidebar. */}
                <div className="mt-auto flex shrink-0 flex-col space-y-0.5 pt-6">
                  <div className="mx-2 mb-2 border-t border-neutral-700" />
                  {footer.map(renderItem)}
                </div>
              </>
            )}
          </nav>

          <ScrollArea
            viewportRef={scrollViewportRef}
            className="box-border h-full w-full min-w-0 rounded-r-lg bg-neutral-900/50 px-3 [&_[data-slot=scroll-area-scrollbar]]:w-1.5 [&_[data-slot=scroll-area-scrollbar]]:p-0 [&_[data-slot=scroll-area-thumb]]:bg-neutral-500/30"
          >
            <div className="flex h-full w-full justify-center">
              <div className="w-full max-w-[600px] pt-14 pb-10">
                <SettingsNav.Provider value={{ openTab: changeTab }}>
                  {tabs.find((tab) => tab.id === currentTab)?.component}
                </SettingsNav.Provider>
              </div>
            </div>
          </ScrollArea>
        </div>
      </Modal>
    </>
  );
};
