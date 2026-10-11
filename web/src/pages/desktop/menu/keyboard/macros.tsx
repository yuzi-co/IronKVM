import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Divider, message, Popconfirm, Popover, Tooltip } from 'antd';
import { useAtomValue, useSetAtom } from 'jotai';
import { CircleIcon, ListVideoIcon, PencilIcon, PlayIcon, Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/sequence.ts';
import { showFailure } from '@/lib/feedback.ts';
import type { Macro } from '@/lib/macro.ts';
import {
  macroDraftAtom,
  macroRecordingAtom,
  macrosVersionAtom,
  sequenceStatusAtom
} from '@/jotai/sequence.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';
import { ScrollArea } from '@/components/ui/scroll-area';
import { startRecording } from '@/pages/desktop/sequence/recording.ts';

const rowClassName =
  'flex h-[30px] w-full items-center space-x-2 rounded px-2 text-left hover:bg-neutral-700/50';

const iconButtonClassName =
  'flex h-6 w-6 shrink-0 cursor-pointer items-center justify-center rounded text-neutral-400 hover:bg-neutral-700/70 hover:text-neutral-100';

// Macros lists the macros saved on the device and records new ones. The server
// plays a macro, so it finishes when this tab closes; the sequence banner
// shows it and stops it.
export const Macros = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';
  const supported = useServerFeature('hid-macros');

  const [isOpen, setIsOpen] = useState(false);
  const [macros, setMacros] = useState<Macro[]>([]);
  const version = useAtomValue(macrosVersionAtom);
  const setVersion = useSetAtom(macrosVersionAtom);
  const setDraft = useSetAtom(macroDraftAtom);
  const setStatus = useSetAtom(sequenceStatusAtom);
  const recording = useAtomValue(macroRecordingAtom);

  useEffect(() => {
    if (!supported || !isOpen) return;
    let live = true;
    api
      .getMacros()
      .then((rsp) => {
        if (!live) return;
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setMacros(rsp.data.macros ?? []);
      })
      .catch((err) => live && showFailure(err));
    return () => {
      live = false;
    };
  }, [supported, isOpen, version]);

  if (!supported) return null;

  function play(macro: Macro) {
    api
      .playMacro({ id: macro.id })
      .then((rsp) => {
        if (rsp.code === -3) {
          message.error(t('keyboard.sequence.alreadyRunning'));
          return;
        }
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setStatus(rsp.data);
        setIsOpen(false);
      })
      .catch((err) => showFailure(err));
  }

  function remove(macro: Macro) {
    api
      .deleteMacro(macro.id)
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setVersion((v) => v + 1);
      })
      .catch((err) => showFailure(err));
  }

  function record() {
    startRecording();
    setIsOpen(false);
  }

  const content = (
    <ScrollArea className="w-[280px] *:data-radix-scroll-area-viewport:max-h-[350px]">
      {macros.length === 0 ? (
        <div className="px-2 py-1.5 text-xs text-neutral-500">{t('keyboard.macro.empty')}</div>
      ) : (
        macros.map((macro) => (
          <div key={macro.id} className={rowClassName}>
            <button
              type="button"
              className="flex min-w-0 flex-1 cursor-pointer items-center space-x-2 p-0 text-left"
              title={t('keyboard.macro.play')}
              onClick={() => play(macro)}
            >
              <PlayIcon size={14} className="shrink-0 text-neutral-400" />
              <span className="truncate">{macro.name}</span>
            </button>
            <Tooltip title={t('keyboard.macro.edit')}>
              <button
                type="button"
                aria-label={t('keyboard.macro.edit')}
                className={iconButtonClassName}
                onClick={() => {
                  setDraft({ id: macro.id, name: macro.name, steps: macro.steps });
                  setIsOpen(false);
                }}
              >
                <PencilIcon size={13} />
              </button>
            </Tooltip>
            {isAdmin && (
              <Popconfirm
                title={t('keyboard.macro.deleteConfirm', { name: macro.name })}
                onConfirm={() => remove(macro)}
              >
                <button
                  type="button"
                  aria-label={t('keyboard.macro.delete')}
                  className={iconButtonClassName}
                >
                  <Trash2Icon size={13} />
                </button>
              </Popconfirm>
            )}
          </div>
        ))
      )}

      <Divider style={{ margin: '5px 0 5px 0' }} />

      <Tooltip title={t('keyboard.macro.recordTip')} placement="right">
        <button
          type="button"
          disabled={recording !== null}
          className={`${rowClassName} cursor-pointer disabled:cursor-not-allowed disabled:text-neutral-500`}
          onClick={record}
        >
          <CircleIcon size={12} className="shrink-0 fill-red-500 text-red-500" />
          <span>{t('keyboard.macro.record')}</span>
        </button>
      </Tooltip>
    </ScrollArea>
  );

  return (
    <Popover
      content={content}
      trigger="hover"
      placement="rightTop"
      align={{ offset: [14, 0] }}
      open={isOpen}
      onOpenChange={setIsOpen}
      arrow={false}
    >
      <div className="flex cursor-pointer items-center space-x-2 rounded py-1 pr-5 pl-2 select-none hover:bg-neutral-700/70">
        <ListVideoIcon size={18} />
        <span>{t('keyboard.macro.title')}</span>
      </div>
    </Popover>
  );
};
