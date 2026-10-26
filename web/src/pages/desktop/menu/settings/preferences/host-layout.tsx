import { useEffect, useState } from 'react';
import { Select } from 'antd';
import { useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/hid.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { layoutLabelKey, pasteLayoutIds } from '@/lib/paste-layout.ts';
import { hostPasteLayoutAtom } from '@/jotai/paste.ts';

// HostLayout is the keyboard layout active on the host. AI agents type with
// it, and the paste dialog starts on it in a browser that has not picked one.
export const HostLayout = () => {
  const { t } = useTranslation();
  const setPasteHostLayout = useSetAtom(hostPasteLayoutAtom);

  const [layout, setLayout] = useState('us');
  // The request starts on mount, so the control begins in its loading state.
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .getHostLayout()
      .then((rsp) => {
        if (rsp.code === 0 && rsp.data?.layout) {
          setLayout(rsp.data.layout);
        }
      })
      .catch(() => {})
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  const options = pasteLayoutIds.map((id) => ({
    value: id,
    label: t(`keyboard.pasting.layouts.${layoutLabelKey(id)}`)
  }));

  function change(value: string) {
    if (isLoading || value === layout) return;
    setIsLoading(true);

    const previous = layout;
    setLayout(value);
    api
      .setHostLayout(value)
      .then((rsp) => {
        if (showResult(rsp, { success: t('feedback.saved') })) {
          setPasteHostLayout({ layout: value, configured: true });
        } else {
          setLayout(previous);
        }
      })
      .catch((err) => {
        setLayout(previous);
        showFailure(err);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="mt-8 flex items-center justify-between space-x-5">
      <div className="flex flex-col">
        <span>{t('settings.appearance.hostLayout')}</span>
        <span className="text-xs text-neutral-500">{t('settings.appearance.hostLayoutDesc')}</span>
      </div>

      <div>
        <Select
          disabled={isLoading}
          showSearch
          optionFilterProp="label"
          style={{ width: 180 }}
          value={layout}
          options={options}
          onChange={change}
        />
      </div>
    </div>
  );
};
