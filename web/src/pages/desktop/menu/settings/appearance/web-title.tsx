import { useEffect, useState } from 'react';
import { Input } from 'antd';
import { useAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { webTitleAtom } from '@/jotai/settings.ts';

export const WebTitle = () => {
  const { t } = useTranslation();
  const [webTitle, setWebTitle] = useAtom(webTitleAtom);

  // The request starts on mount, so the control begins in its loading state.
  // Setting the flag inside the effect left one paint where the control was
  // interactive and the value behind it was not yet known.
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .getWebTitle()
      .then((rsp) => {
        if (rsp.data?.title) {
          setWebTitle(rsp.data.title);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  function submit() {
    if (isLoading) return;
    setIsLoading(true);

    api
      .setWebTitle(webTitle)
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }
      })
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
          placeholder="NanoKVM"
        />
      </div>
    </div>
  );
};
