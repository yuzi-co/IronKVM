import { useEffect, useState } from 'react';
import { Button, Input, message } from 'antd';
import { CheckIcon, ClipboardPenIcon, XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { isValidHostname } from '@/lib/hostname.ts';

export const Hostname = ({ editable = false }: { editable?: boolean }) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(true);
  const [hostname, setHostname] = useState('');

  const [editState, setEditState] = useState<'' | 'editing' | 'edited'>('');
  const [input, setInput] = useState('');
  const [errMsg, setErrMsg] = useState('');

  useEffect(() => {
    api
      .getHostname()
      .then((rsp) => {
        if (rsp.data?.hostname) {
          setHostname(rsp.data?.hostname);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  function showInput() {
    setInput(hostname);
    setErrMsg('');
    setEditState('editing');
  }

  const name = input.trim();
  const isValid = isValidHostname(name);

  function update() {
    if (name === hostname) {
      setEditState('');
      return;
    }

    if (isLoading || !isValid) return;
    setIsLoading(true);
    setErrMsg('');

    api
      .setHostname(name)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(describeFailure(rsp, t('settings.about.hostnameFailed')));
          return;
        }

        setHostname(name);
        setEditState('edited');
        message.success(t('feedback.saved'));
      })
      .catch((err) => setErrMsg(describeFailure(err, t('settings.about.hostnameFailed'))))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="space-y-1">
      <div className="flex w-full items-center justify-between">
        <span>{t('settings.about.hostname')}</span>

        {editState === 'editing' ? (
          <div className="flex items-center space-x-1">
            <Input
              disabled={isLoading}
              style={{ width: 150 }}
              value={input}
              maxLength={64}
              status={isValid ? undefined : 'error'}
              onChange={(e) => setInput(e.target.value)}
              onPressEnter={update}
            />
            <Button
              size="small"
              aria-label={t('common.save')}
              icon={<CheckIcon size={15} />}
              disabled={!isValid}
              loading={isLoading}
              onClick={update}
            />
            <Button
              size="small"
              aria-label={t('common.cancel')}
              icon={<XIcon size={15} />}
              onClick={() => setEditState('')}
            />
          </div>
        ) : (
          <div className="flex items-center space-x-2">
            <span>{hostname}</span>
            {editable && (
              <button
                type="button"
                aria-label={t('settings.about.editHostname')}
                className="size-[16px] cursor-pointer p-0 text-neutral-500 hover:text-blue-500"
                onClick={showInput}
              >
                <ClipboardPenIcon size={16} />
              </button>
            )}
          </div>
        )}
      </div>

      {editState === 'editing' && !isValid && (
        <div className="flex w-full justify-end text-xs text-red-500">
          {t('settings.about.hostnameInvalid')}
        </div>
      )}

      {errMsg && <div className="flex w-full justify-end text-xs text-red-500">{errMsg}</div>}

      {editState === 'edited' && (
        <div className="flex w-full justify-end text-xs text-green-500">
          {t('settings.about.hostnameUpdated')}
        </div>
      )}
    </div>
  );
};
