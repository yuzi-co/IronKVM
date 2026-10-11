import { useEffect, useState, type ReactNode } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, Input, message, Modal, Tooltip } from 'antd';
import type { TFunction } from 'i18next';
import { useAtom, useSetAtom } from 'jotai';
import {
  ClockIcon,
  KeyboardIcon,
  MousePointerIcon,
  PlusIcon,
  PowerIcon,
  RotateCcwIcon,
  Trash2Icon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/sequence.ts';
import { showFailure } from '@/lib/feedback.ts';
import {
  comboLabel,
  DEFAULT_ATX_MS,
  durationMs,
  MAX_DELAY_MS,
  MAX_NAME_LENGTH,
  MAX_STEPS,
  type MacroStep,
  type MouseStep
} from '@/lib/macro.ts';
import { keyboardLockAtom } from '@/jotai/keyboard.ts';
import {
  macroDraftAtom,
  macrosVersionAtom,
  sequenceStatusAtom,
  type MacroDraft
} from '@/jotai/sequence.ts';
import { ScrollArea } from '@/components/ui/scroll-area.tsx';

// The pause a new pause step starts with.
const newPauseMs = 500;

// Absolute pointer positions run from 0 to 32767 across the screen.
const absoluteRange = 32767;

function buttonNames(buttons: number, t: TFunction): string {
  const names: string[] = [];
  if (buttons & 1) names.push(t('keyboard.macro.buttonLeft'));
  if (buttons & 2) names.push(t('keyboard.macro.buttonRight'));
  if (buttons & 4) names.push(t('keyboard.macro.buttonMiddle'));
  return names.join(', ');
}

function describeMouse(step: MouseStep, t: TFunction): string {
  const parts: string[] = [];
  const x = step.x ?? 0;
  const y = step.y ?? 0;
  if (step.absolute) {
    parts.push(
      t('keyboard.macro.stepPointerAt', {
        x: Math.round((x / absoluteRange) * 100),
        y: Math.round((y / absoluteRange) * 100)
      })
    );
  } else {
    parts.push(t('keyboard.macro.stepPointerBy', { x, y }));
  }
  parts.push(
    step.buttons
      ? t('keyboard.macro.stepButtons', { buttons: buttonNames(step.buttons, t) })
      : t('keyboard.macro.stepNoButtons')
  );
  const scroll = step.wheel || step.hwheel;
  if (scroll) parts.push(t('keyboard.macro.stepScroll', { amount: scroll }));
  return parts.join(', ');
}

// MacroEditor opens when a recording finishes or a saved macro is edited. The
// steps can be removed, pauses changed or added, and ATX presses added; the
// macro can be played before it is saved.
export const MacroEditor = () => {
  const [draft, setDraft] = useAtom(macroDraftAtom);
  if (!draft) return null;
  return <EditorModal draft={draft} onClose={() => setDraft(null)} />;
};

type EditorModalProps = {
  draft: MacroDraft;
  onClose: () => void;
};

const EditorModal = ({ draft, onClose }: EditorModalProps) => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';
  const setKeyboardLock = useSetAtom(keyboardLockAtom);
  const setStatus = useSetAtom(sequenceStatusAtom);
  const setMacrosVersion = useSetAtom(macrosVersionAtom);

  const [name, setName] = useState(draft.name);
  const [steps, setSteps] = useState<MacroStep[]>(draft.steps);
  const [isSaving, setIsSaving] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);

  // Typing the name must not type on the host.
  useEffect(() => {
    setKeyboardLock({ source: 'macro-editor', locked: true });
    return () => setKeyboardLock({ source: 'macro-editor', locked: false });
  }, [setKeyboardLock]);

  function setPause(index: number, value: string) {
    const ms = Math.min(MAX_DELAY_MS, Math.max(0, Math.round(Number(value) || 0)));
    setSteps((current) => current.map((step, i) => (i === index ? { type: 'delay', ms } : step)));
  }

  function remove(index: number) {
    setSteps((current) => current.filter((_, i) => i !== index));
  }

  function add(step: MacroStep) {
    setSteps((current) => (current.length >= MAX_STEPS ? current : [...current, step]));
  }

  function check(): boolean {
    if (steps.length === 0) {
      message.error(t('keyboard.macro.noSteps'));
      return false;
    }
    return true;
  }

  function play() {
    if (isPlaying || !check()) return;
    setIsPlaying(true);
    api
      .playMacro({ name: name.trim() || t('keyboard.macro.editorTitle'), steps })
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
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsPlaying(false));
  }

  function save() {
    if (isSaving || !check()) return;
    if (!name.trim()) {
      message.error(t('keyboard.macro.noName'));
      return;
    }
    setIsSaving(true);
    api
      .saveMacro({ id: draft.id, name: name.trim(), steps })
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        message.success(t('keyboard.macro.saved'));
        setMacrosVersion((v) => v + 1);
        onClose();
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsSaving(false));
  }

  function renderStep(step: MacroStep, index: number) {
    let icon = <KeyboardIcon size={14} />;
    let body: ReactNode = null;

    switch (step.type) {
      case 'key': {
        const combo = comboLabel(step.modifiers, step.keys);
        body = combo
          ? t('keyboard.macro.stepKeys', { keys: combo })
          : t('keyboard.macro.stepRelease');
        break;
      }
      case 'mouse':
        icon = <MousePointerIcon size={14} />;
        body = describeMouse(step, t);
        break;
      case 'delay':
        icon = <ClockIcon size={14} />;
        body = (
          <label className="flex items-center space-x-2">
            <span>{t('keyboard.macro.stepPause')}</span>
            <input
              type="number"
              min={0}
              max={MAX_DELAY_MS}
              step={10}
              value={step.ms}
              onChange={(e) => setPause(index, e.target.value)}
              className="w-24 rounded border border-neutral-700 bg-neutral-800 px-1.5 py-0.5 text-neutral-200"
            />
          </label>
        );
        break;
      case 'atx':
        icon = step.button === 'power' ? <PowerIcon size={14} /> : <RotateCcwIcon size={14} />;
        body = t(
          step.button === 'power' ? 'keyboard.macro.stepPower' : 'keyboard.macro.stepReset',
          { ms: step.ms || DEFAULT_ATX_MS }
        );
        break;
    }

    return (
      <div
        key={index}
        className="flex min-h-[30px] items-center space-x-2 rounded px-2 text-sm text-neutral-300 hover:bg-neutral-700/40"
      >
        <span className="w-8 shrink-0 text-right text-xs text-neutral-500">{index + 1}</span>
        <span className="shrink-0 text-neutral-400">{icon}</span>
        <div className="min-w-0 flex-1 truncate">{body}</div>
        <Tooltip title={t('keyboard.macro.removeStep')} placement="left">
          <button
            type="button"
            aria-label={t('keyboard.macro.removeStep')}
            className="flex h-6 w-6 shrink-0 cursor-pointer items-center justify-center rounded text-neutral-500 hover:bg-neutral-700/70 hover:text-red-400"
            onClick={() => remove(index)}
          >
            <Trash2Icon size={14} />
          </button>
        </Tooltip>
      </div>
    );
  }

  const saveButton = (
    <Button type="primary" loading={isSaving} disabled={!isAdmin} onClick={save}>
      {t('keyboard.macro.save')}
    </Button>
  );

  return (
    <Modal
      open
      title={t('keyboard.macro.editorTitle')}
      width={640}
      onCancel={onClose}
      maskClosable={false}
      footer={[
        <Button key="cancel" onClick={onClose}>
          {t('keyboard.macro.discard')}
        </Button>,
        <Button key="play" loading={isPlaying} onClick={play}>
          {t('keyboard.macro.playNow')}
        </Button>,
        isAdmin ? (
          <span key="save">{saveButton}</span>
        ) : (
          <Tooltip key="save" title={t('keyboard.macro.adminOnly')}>
            {saveButton}
          </Tooltip>
        )
      ]}
    >
      <div className="flex flex-col space-y-3">
        <Input
          value={name}
          maxLength={MAX_NAME_LENGTH}
          placeholder={t('keyboard.macro.namePlaceholder')}
          aria-label={t('keyboard.macro.name')}
          onChange={(e) => setName(e.target.value)}
        />

        <div className="text-xs text-neutral-400">
          {t('keyboard.macro.summary', {
            count: steps.length,
            seconds: Math.round(durationMs(steps) / 100) / 10
          })}
        </div>

        <ScrollArea className="rounded border border-neutral-800 *:data-radix-scroll-area-viewport:max-h-[50vh]">
          <div className="py-1">
            {steps.length === 0 ? (
              <div className="px-3 py-2 text-sm text-neutral-500">
                {t('keyboard.macro.noSteps')}
              </div>
            ) : (
              steps.map(renderStep)
            )}
          </div>
        </ScrollArea>

        <div className="flex flex-wrap gap-2">
          <Button
            size="small"
            icon={<PlusIcon size={14} />}
            onClick={() => add({ type: 'delay', ms: newPauseMs })}
          >
            {t('keyboard.macro.addPause')}
          </Button>
          <Button
            size="small"
            icon={<PowerIcon size={14} />}
            onClick={() => add({ type: 'atx', button: 'power', ms: DEFAULT_ATX_MS })}
          >
            {t('keyboard.macro.addPower')}
          </Button>
          <Button
            size="small"
            icon={<RotateCcwIcon size={14} />}
            onClick={() => add({ type: 'atx', button: 'reset', ms: DEFAULT_ATX_MS })}
          >
            {t('keyboard.macro.addReset')}
          </Button>
        </div>
      </div>
    </Modal>
  );
};
