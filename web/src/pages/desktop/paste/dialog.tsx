import { ChangeEvent, useEffect, useRef, useState } from 'react';
import { Button, Divider, Input, Modal, Select } from 'antd';
import type { TextAreaRef } from 'antd/es/input/TextArea';
import { useAtom } from 'jotai';
import { ClipboardPasteIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { checkPaste, type PasteCheck, type PasteUntypeable } from '@/api/hid.ts';
import { readClipboardText } from '@/lib/clipboard.ts';
import {
  defaultPasteDelay,
  pasteDelayAtom,
  pasteDialogAtom,
  pasteLayoutAtom,
  pasteLayoutIds,
  pasteTextAtom
} from '@/jotai/paste.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

import {
  formatDuration,
  maxPasteLength,
  pasteLength,
  pasteShortcutLabel,
  usePaste
} from './use-paste.ts';

// Fixed for the lifetime of the document: the Clipboard API is there in a
// secure context and absent outside one.
const canReadClipboard = window.isSecureContext === true && !!navigator.clipboard?.readText;

// How many untypeable characters the dialog lists. The server reports up to a
// hundred, and the count covers the rest.
const untypeableShown = 20;

const layoutLabelKeys: Record<string, string> = { 'pt-br': 'ptBr' };

function codePoint(char: string) {
  return Array.from(char)
    .map((c) => 'U+' + (c.codePointAt(0) ?? 0).toString(16).toUpperCase().padStart(4, '0'))
    .join(' ');
}

type LiveCheck = { key: string; check: PasteCheck };

export const PasteDialog = () => {
  const { t } = useTranslation();

  const [dialog, setDialog] = useAtom(pasteDialogAtom);
  useKeyboardLock('paste-modal', dialog.open);
  const [text, setText] = useAtom(pasteTextAtom);
  const [layout, setLayout] = useAtom(pasteLayoutAtom);
  const [delay, setDelay] = useAtom(pasteDelayAtom);
  const { start } = usePaste();

  const [liveCheck, setLiveCheck] = useState<LiveCheck | null>(null);
  const [isReadingClipboard, setIsReadingClipboard] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  const textAreaRef = useRef<TextAreaRef>(null);
  const checkSeq = useRef(0);

  const length = pasteLength(text);
  const isTooLong = length > maxPasteLength;
  const checkKey = `${layout}|${delay}|${text}`;

  // The server's report on the text as it stands. Until the check of the
  // latest edit arrives, a report the server sent with a refusal stands in.
  const check =
    !text || isTooLong ? null : liveCheck?.key === checkKey ? liveCheck.check : dialog.check;
  const untypeableCount = check?.untypeableCount ?? 0;

  useEffect(() => {
    if (!dialog.open || !text || isTooLong) return;

    const key = `${layout}|${delay}|${text}`;
    const timer = setTimeout(() => {
      const seq = ++checkSeq.current;
      checkPaste(text, layout, delay)
        .then((rsp) => {
          if (rsp.code === 0 && seq === checkSeq.current) {
            setLiveCheck({ key, check: rsp.data });
          }
        })
        .catch(() => {
          // The paste itself reports what the check would have.
        });
    }, 300);

    return () => clearTimeout(timer);
  }, [dialog.open, text, layout, delay, isTooLong]);

  const layouts = pasteLayoutIds.map((id) => ({
    value: id,
    label: t(`keyboard.pasting.layouts.${layoutLabelKeys[id] ?? id}`)
  }));

  const speeds = [
    { value: 15, label: t('keyboard.pasting.speeds.fast') },
    { value: defaultPasteDelay, label: t('keyboard.pasting.speeds.normal') },
    { value: 100, label: t('keyboard.pasting.speeds.slow') }
  ];

  function onChange(e: ChangeEvent<HTMLTextAreaElement>) {
    setText(e.target.value);
  }

  function close() {
    setDialog({ open: false, notice: '', check: null });
  }

  async function readFromClipboard() {
    if (isReadingClipboard) return;
    setIsReadingClipboard(true);

    const result = await readClipboardText();
    setIsReadingClipboard(false);

    if (!result.ok) {
      setErrMsg(
        result.reason === 'unavailable'
          ? t('keyboard.pasting.clipboardUnavailable')
          : result.reason === 'denied'
            ? t('keyboard.clipboardPermissionDenied')
            : t('keyboard.clipboardReadError')
      );
      return;
    }
    if (!result.text) {
      setErrMsg(t('keyboard.pasting.clipboardEmpty'));
      return;
    }
    setText((value) => value + result.text);
  }

  // select marks one untypeable character in the text. The server counts code
  // points and the text area UTF-16 units, so the offset is converted.
  function select(item: PasteUntypeable) {
    const textArea = textAreaRef.current?.resizableTextArea?.textArea;
    if (!textArea) return;

    const from = Array.from(text).slice(0, item.index).join('').length;
    textArea.focus();
    textArea.setSelectionRange(from, from + item.char.length);
  }

  async function submit() {
    if (isStarting || !text || isTooLong) return;
    setIsStarting(true);
    setErrMsg('');

    const started = await start(text, untypeableCount > 0);
    setIsStarting(false);
    if (started) {
      setText('');
      close();
    }
  }

  function afterOpenChange(open: boolean) {
    if (open) {
      textAreaRef.current?.focus();
    }
  }

  const notice = errMsg || dialog.notice;

  return (
    <Modal
      open={dialog.open}
      centered={false}
      footer={null}
      onCancel={close}
      afterOpenChange={afterOpenChange}
    >
      <div className="flex flex-col">
        <span className="text-xl">{t('keyboard.paste')}</span>
        <span className="text-sm text-neutral-400">{t('keyboard.tips')}</span>
      </div>

      <Divider style={{ margin: '14px 0' }} />

      <div className="flex w-full flex-wrap items-center gap-3 pb-2">
        {canReadClipboard && (
          <Button
            color="default"
            variant="filled"
            icon={<ClipboardPasteIcon size={16} />}
            loading={isReadingClipboard}
            onClick={readFromClipboard}
            className="flex items-center"
          >
            {t('keyboard.readClipboard')}
          </Button>
        )}

        <Select
          value={layout}
          variant="filled"
          onChange={setLayout}
          options={layouts}
          popupMatchSelectWidth={false}
          title={t('keyboard.pasting.layout')}
          aria-label={t('keyboard.pasting.layout')}
        />

        <Select
          value={delay}
          variant="filled"
          onChange={setDelay}
          options={speeds}
          popupMatchSelectWidth={false}
          title={t('keyboard.pasting.speed')}
          aria-label={t('keyboard.pasting.speed')}
        />
      </div>

      <Input.TextArea
        ref={textAreaRef}
        value={text}
        status={isTooLong || untypeableCount > 0 ? 'error' : ''}
        showCount
        maxLength={maxPasteLength}
        autoSize={{ minRows: 6, maxRows: 12 }}
        placeholder={t('keyboard.placeholder')}
        onFocus={() => setErrMsg('')}
        onChange={onChange}
      />

      {notice && <div className="pt-1 text-sm text-red-500">{notice}</div>}

      {check && check.untypeableCount > 0 && (
        <div className="pt-2 text-sm">
          <div className="text-red-500">
            {t('keyboard.pasting.untypeable', { count: check.untypeableCount })}
          </div>
          <div className="flex flex-wrap gap-1 pt-1">
            {check.untypeable.slice(0, untypeableShown).map((item) => (
              <button
                key={item.index}
                type="button"
                className="cursor-pointer rounded bg-neutral-700/40 px-1.5 py-0.5 text-xs hover:bg-neutral-700/70"
                title={codePoint(item.char)}
                onClick={() => select(item)}
              >
                <span className="font-mono">
                  {item.char.trim() ? item.char : codePoint(item.char)}
                </span>
                <span className="pl-1 text-neutral-500">
                  {t('keyboard.pasting.untypeableAt', { line: item.line, column: item.column })}
                </span>
              </button>
            ))}
            {check.untypeableCount > untypeableShown && (
              <span className="px-1 text-xs text-neutral-500">
                +{check.untypeableCount - untypeableShown}
              </span>
            )}
          </div>
        </div>
      )}

      {check && check.characters > 0 && (
        <div className="pt-2 text-xs text-neutral-500">
          {t('keyboard.pasting.estimate', { duration: formatDuration(check.durationMs) })}
        </div>
      )}

      <div className="flex justify-center pt-5 pb-2">
        <Button
          type="primary"
          loading={isStarting}
          disabled={!text || isTooLong || check?.characters === 0}
          htmlType="submit"
          style={{ width: '300px' }}
          onClick={submit}
        >
          {untypeableCount > 0 ? t('keyboard.pasting.skipUntypeable') : t('keyboard.submit')}
        </Button>
      </div>

      <div className="text-center text-xs text-neutral-500">
        {t('keyboard.pasting.shortcut', { shortcut: pasteShortcutLabel })}
      </div>
    </Modal>
  );
};
