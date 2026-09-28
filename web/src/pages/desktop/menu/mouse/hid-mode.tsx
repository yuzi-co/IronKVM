import { useState } from 'react';
import { Button, Divider, Modal, Typography } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { PenIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/hid.ts';
import { refreshHidModeAtom } from '@/jotai/hid.ts';
import { useHidMode } from '@/hooks/useHidMode.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

const { Paragraph } = Typography;

export const HidMode = () => {
  const { t } = useTranslation();

  const current = useHidMode();
  const refreshHidMode = useSetAtom(refreshHidModeAtom);
  const [isSwitching, setIsSwitching] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  useKeyboardLock('hid-mode-modal', isModalOpen);

  const hidMode = current?.mode ?? 'normal';
  const isLoading = current === null || isSwitching;

  // A first fetch that failed leaves the mode unknown, and the switch cannot
  // offer the right direction. Opening the dialog asks again.
  function openModal() {
    if (current === null) {
      refreshHidMode();
    }
    setIsModalOpen(true);
  }

  // The switch used to reboot the board, so this waited thirty seconds and then
  // reloaded the page. It rebuilds the USB gadget instead, which takes about a
  // second and leaves the network, the video and this session alone, so the
  // answer can simply be believed. The media and power keys follow the mode,
  // so the shared copy is fetched again for every menu that shows them.
  function updateHidMode() {
    if (isLoading) return;
    setIsSwitching(true);
    setErrMsg('');

    const mode = hidMode === 'normal' ? 'hid-only' : 'normal';

    api
      .setHidMode(mode)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }

        setIsModalOpen(false);
        return refreshHidMode(true);
      })
      .catch((err) => {
        console.log(err);
        setErrMsg(t('mouse.hidOnly.switchFailed'));
      })
      .finally(() => {
        setIsSwitching(false);
      });
  }

  return (
    <>
      <button
        type="button"
        className={clsx(
          'flex h-[30px] w-full cursor-pointer items-center space-x-2 rounded p-0 px-3 text-left select-none hover:bg-neutral-700/70',
          hidMode === 'hid-only' ? 'text-blue-500' : 'text-neutral-300'
        )}
        onClick={openModal}
      >
        <PenIcon size={18} />
        <span>{t('mouse.hidOnly.title')}</span>
      </button>

      <Modal
        open={isModalOpen}
        title={t('mouse.hidOnly.title')}
        width={580}
        centered={false}
        footer={false}
        onCancel={() => setIsModalOpen(false)}
      >
        <Divider />

        <Paragraph>{t('mouse.hidOnly.desc')}</Paragraph>

        <Paragraph type="secondary">
          <ul>
            <li>{t('mouse.hidOnly.tip1')}</li>
            <li>{t('mouse.hidOnly.tip2')}</li>
            <li>{t('mouse.hidOnly.rebuild')}</li>
          </ul>
        </Paragraph>

        {errMsg && <div className="pt-1 text-sm text-red-500">{errMsg}</div>}

        <div className="flex justify-center pt-5">
          <Button danger type="primary" loading={isLoading} onClick={updateHidMode}>
            {hidMode === 'normal' ? t('mouse.hidOnly.enable') : t('mouse.hidOnly.disable')}
          </Button>
        </div>
      </Modal>
    </>
  );
};
