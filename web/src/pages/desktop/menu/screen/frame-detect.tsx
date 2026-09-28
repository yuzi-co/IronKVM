import { useEffect, useState } from 'react';
import { message, Tooltip } from 'antd';
import clsx from 'clsx';
import { LoaderCircleIcon, Tally4Icon, Tally5Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/stream.ts';

export const FrameDetect = () => {
  const { t } = useTranslation();

  // The setting is board-wide, so it is read from the board rather than kept
  // in this browser. It stays loading until the answer comes.
  const [isLoading, setIsLoading] = useState(true);
  const [isEnabled, setIsEnabled] = useState(false);

  useEffect(() => {
    api
      .getFrameDetect()
      .then((rsp) => {
        if (rsp.code === 0) setIsEnabled(!!rsp.data?.enabled);
      })
      .catch(() => {})
      .finally(() => setIsLoading(false));
  }, []);

  function update() {
    if (isLoading) return;
    setIsLoading(true);

    const enabled = !isEnabled;

    api
      .updateFrameDetect(enabled)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('screen.updateFailed'));
          return;
        }
        setIsEnabled(enabled);
      })
      .catch(() => message.error(t('screen.updateFailed')))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <Tooltip placement="rightTop" title={t('screen.frameDetectTip')} color="#262626" arrow>
      <div
        className="group flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700"
        onClick={update}
      >
        {isLoading ? (
          <LoaderCircleIcon className="animate-spin" size={18} />
        ) : (
          <>
            {isEnabled ? <Tally4Icon color="#22c55e" size={18} /> : <Tally5Icon size={18} />}

            <span
              className={clsx(
                'text-sm select-none',
                isEnabled ? 'group-hover:text-red-500' : 'group-hover:text-green-500'
              )}
            >
              {t('screen.frameDetect')}
            </span>
          </>
        )}
      </div>
    </Tooltip>
  );
};
