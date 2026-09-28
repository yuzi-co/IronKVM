import { useEffect, useState } from 'react';
import { CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { Button, Input } from 'antd';
import { ClipboardPenIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
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
          setErrMsg(rsp.msg || t('settings.about.hostnameFailed'));
          return;
        }

        setHostname(name);
        setEditState('edited');
      })
      .catch(() => setErrMsg(t('settings.about.hostnameFailed')))
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
              icon={<CheckOutlined />}
              disabled={!isValid}
              loading={isLoading}
              onClick={update}
            />
            <Button size="small" icon={<CloseOutlined />} onClick={() => setEditState('')} />
          </div>
        ) : (
          <div className="flex items-center space-x-2">
            <span>{hostname}</span>
            {editable && (
              <div
                className="size-[16px] cursor-pointer text-neutral-500 hover:text-blue-500"
                onClick={showInput}
              >
                <ClipboardPenIcon size={16} />
              </div>
            )}
          </div>
        )}
      </div>

      {editState === 'editing' && !isValid && (
        <div className="flex w-full justify-end text-xs text-red-500">
          {t('settings.about.hostnameInvalid')}
        </div>
      )}

      {errMsg && (
        <div className="flex w-full justify-end text-xs text-red-500">{errMsg}</div>
      )}

      {editState === 'edited' && (
        <div className="flex w-full justify-end text-xs text-green-500">
          {t('settings.about.hostnameUpdated')}
        </div>
      )}
    </div>
  );
};
