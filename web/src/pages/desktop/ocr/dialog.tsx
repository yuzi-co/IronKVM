import { useEffect, useMemo } from 'react';
import { Alert, Button, Input, message, Modal, Progress, Select, Spin } from 'antd';
import { CopyIcon, ScanTextIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { writeClipboardText } from '@/lib/clipboard.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

import { ocrLanguages, type OcrLanguage } from './engine.ts';

export type OcrError = 'unsupported' | 'capture' | 'outside' | 'recognize';

export type OcrState =
  | { phase: 'capturing' }
  | { phase: 'loading' | 'recognizing'; progress: number }
  | { phase: 'done'; text: string }
  | { phase: 'error'; error: OcrError; detail: string };

const errorKeys: Record<OcrError, string> = {
  unsupported: 'screen.ocr.unsupported',
  capture: 'screen.ocr.captureFailed',
  outside: 'screen.ocr.outside',
  recognize: 'screen.ocr.recognizeFailed'
};

type OcrDialogProps = {
  open: boolean;
  state: OcrState;
  // The crop that was read, as the engine saw it.
  image: Blob | null;
  language: OcrLanguage;
  onLanguageChange: (language: OcrLanguage) => void;
  onTextChange: (text: string) => void;
  onSelectAgain: () => void;
  onClose: () => void;
};

export const OcrDialog = ({
  open,
  state,
  image,
  language,
  onLanguageChange,
  onTextChange,
  onSelectAgain,
  onClose
}: OcrDialogProps) => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();

  const isBusy =
    state.phase === 'capturing' || state.phase === 'loading' || state.phase === 'recognizing';

  // The lock follows open rather than the end of the opening animation, so no
  // key reaches the host between the selection closing and the dialog opening.
  useKeyboardLock('ocr-dialog', open);

  const previewUrl = useMemo(() => (image ? URL.createObjectURL(image) : ''), [image]);
  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  const languages = ocrLanguages.map((code) => ({
    value: code,
    label: t(`screen.ocr.languages.${code}`)
  }));

  async function copy() {
    if (state.phase !== 'done') return;

    try {
      await writeClipboardText(state.text);
      messageApi.success(t('screen.ocr.copied'));
    } catch {
      messageApi.error(t('screen.ocr.copyFailed'));
    }
  }

  function status() {
    switch (state.phase) {
      case 'capturing':
        return (
          <div className="flex items-center gap-3 py-6">
            <Spin size="small" />
            <span>{t('screen.ocr.capturing')}</span>
          </div>
        );
      case 'loading':
      case 'recognizing':
        return (
          <div className="py-4">
            <div className="pb-1">
              {state.phase === 'loading' ? t('screen.ocr.loading') : t('screen.ocr.recognizing')}
            </div>
            <Progress percent={Math.round(state.progress * 100)} size="small" />
          </div>
        );
      case 'error':
        return (
          <Alert
            type="error"
            showIcon
            message={t(errorKeys[state.error])}
            description={state.detail || undefined}
          />
        );
      case 'done':
        return (
          <Input.TextArea
            value={state.text}
            autoSize={{ minRows: 6, maxRows: 16 }}
            placeholder={t('screen.ocr.noText')}
            className="font-mono"
            onChange={(event) => onTextChange(event.target.value)}
          />
        );
    }
  }

  return (
    <Modal open={open} centered={false} footer={null} onCancel={onClose} width={640}>
      {contextHolder}

      <div className="flex flex-col pb-3">
        <span className="text-base font-bold text-neutral-300">{t('screen.ocr.title')}</span>
        <span className="text-xs text-neutral-500">{t('screen.ocr.tips')}</span>
      </div>

      <div className="flex flex-wrap items-center gap-3 pb-3">
        <Select
          value={language}
          variant="filled"
          disabled={isBusy}
          onChange={onLanguageChange}
          options={languages}
          popupMatchSelectWidth={false}
          title={t('screen.ocr.language')}
          aria-label={t('screen.ocr.language')}
        />
      </div>

      {previewUrl && (
        <div className="flex justify-center pb-3">
          <img
            src={previewUrl}
            alt={t('screen.ocr.preview')}
            className="max-h-32 max-w-full rounded border border-neutral-700 object-contain"
          />
        </div>
      )}

      {status()}

      <div className="flex flex-wrap justify-end gap-3 pt-4">
        <Button icon={<ScanTextIcon size={16} />} onClick={onSelectAgain}>
          {t('screen.ocr.selectAgain')}
        </Button>
        <Button
          type="primary"
          icon={<CopyIcon size={15} />}
          disabled={state.phase !== 'done' || !state.text}
          onClick={copy}
        >
          {t('screen.ocr.copy')}
        </Button>
      </div>
    </Modal>
  );
};
