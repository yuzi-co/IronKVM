import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { Button, message, Tooltip } from 'antd';
import { CheckIcon, CopyIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { writeClipboardText } from '@/lib/clipboard.ts';

type CopyButtonProps = {
  text: string;
  // What is copied, for the tooltip and screen readers.
  label?: string;
};

// CopyButton copies text and shows a check for two seconds. The clipboard
// helper falls back to execCommand, which a board on plain http needs.
export const CopyButton = ({ text, label }: CopyButtonProps) => {
  const { t } = useTranslation();
  const [isCopied, setIsCopied] = useState(false);

  useEffect(() => {
    if (!isCopied) return;
    const timer = window.setTimeout(() => setIsCopied(false), 2000);
    return () => window.clearTimeout(timer);
  }, [isCopied]);

  async function copy() {
    try {
      await writeClipboardText(text);
      setIsCopied(true);
    } catch {
      message.error(t('common.copyFailed'));
    }
  }

  const title = isCopied
    ? t('common.copied')
    : label
      ? `${t('common.copy')}: ${label}`
      : t('common.copy');

  return (
    <Tooltip title={title}>
      <Button
        type="text"
        size="small"
        className="shrink-0 text-neutral-400 hover:text-white"
        aria-label={title}
        disabled={!text}
        icon={
          isCopied ? <CheckIcon size={15} className="text-green-500" /> : <CopyIcon size={15} />
        }
        onClick={copy}
      />
    </Tooltip>
  );
};

type CopyRowProps = {
  label: ReactNode;
  value: string;
  // Shown instead of the value, when it should be a link.
  display?: ReactNode;
  copyLabel?: string;
};

// CopyRow is a label on the left and a copyable monospace value on the right.
export const CopyRow = ({ label, value, display, copyLabel }: CopyRowProps) => (
  <div className="flex flex-wrap items-center justify-between gap-x-3">
    <span className="text-neutral-400">{label}</span>
    <div className="flex min-w-0 items-center gap-1">
      <span className="min-w-0 font-mono text-xs break-all text-neutral-300 select-all">
        {display ?? value}
      </span>
      {value && <CopyButton text={value} label={copyLabel} />}
    </div>
  </div>
);

type CopyBlockProps = {
  title?: ReactNode;
  text: string;
  copyLabel?: string;
};

// CopyBlock is a command or config snippet in a box, with a copy button.
export const CopyBlock = ({ title, text, copyLabel }: CopyBlockProps) => (
  <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3.5">
    {title && <span className="text-sm font-medium text-neutral-400">{title}</span>}
    <div className="flex min-w-0 items-start justify-between gap-2">
      <pre className="m-0 min-w-0 flex-1 font-mono text-sm break-all whitespace-pre-wrap text-neutral-300 select-all">
        {text}
      </pre>
      <CopyButton text={text} label={copyLabel} />
    </div>
  </div>
);
