import { useEffect, useRef, useState } from 'react';
import { Input } from 'antd';
import { useAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { webTitleAtom } from '@/jotai/settings.ts';

export const WebTitle = () => {
  const { t } = useTranslation();
  const [webTitle, setWebTitle] = useAtom(webTitleAtom);

  // The request starts on mount, so the control begins in its loading state.
  // Setting the flag inside the effect left one paint where the control was
  // interactive and the value behind it was not yet known.
  const [isLoading, setIsLoading] = useState(true);
  // The title the server holds. Enter and then blur both submit, and neither
  // should send a title that has not changed.
  const saved = useRef('');

  useEffect(() => {
    api
      .getWebTitle()
      .then((rsp) => {
        if (rsp.data?.title) {
          setWebTitle(rsp.data.title);
          saved.current = rsp.data.title;
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [setWebTitle]);

  function submit() {
    if (isLoading || webTitle === saved.current) return;
    setIsLoading(true);

    const title = webTitle;
    api
      .setWebTitle(title)
      .then((rsp) => {
        if (showResult(rsp, { success: t('feedback.saved') })) {
          saved.current = title;
        }
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="mt-8 flex items-center justify-between space-x-5">
      <div className="flex flex-col">
        <span>{t('settings.appearance.webTitle')}</span>
        <span className="text-xs text-neutral-500">{t('settings.appearance.webTitleDesc')}</span>
      </div>

      <div>
        <Input
          disabled={isLoading}
          style={{ width: 180 }}
          value={webTitle}
          onChange={(e) => setWebTitle(e.target.value)}
          onPressEnter={submit}
          onBlur={submit}
          placeholder="IronKVM"
        />
      </div>
    </div>
  );
};
