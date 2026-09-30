import { useState } from 'react';
import { Modal } from 'antd';
import { Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

import { ErrorDetail } from './error-detail.tsx';
import { MenuRow } from './menu-row.tsx';
import type { VpnInfo } from './types.ts';

type UninstallProps = {
  vpn: VpnInfo;
  onSuccess: () => void;
};

export const Uninstall = ({ vpn, onSuccess }: UninstallProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  function uninstall() {
    if (isLoading) return;
    setIsLoading(true);

    setErrMsg('');

    // A refused uninstall keeps the dialog open with the reason in it.
    vpn.api
      .uninstall()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg || t('settings.vpn.uninstallFailed'));
          return;
        }
        setIsModalOpen(false);
        onSuccess();
      })
      .catch((err) => setErrMsg(describeFailure(err, t('settings.vpn.uninstallFailed'))))
      .finally(() => setIsLoading(false));
  }

  const title = (
    <div className="flex items-center space-x-1 text-red-500">
      <Trash2Icon size={18} />
      <span>{t('settings.vpn.uninstall', { name: vpn.title })}</span>
    </div>
  );

  return (
    <>
      <MenuRow
        icon={<Trash2Icon size={18} />}
        label={t('settings.vpn.uninstall', { name: vpn.title })}
        onClick={() => {
          setErrMsg('');
          setIsModalOpen(true);
        }}
      />

      <Modal
        title={title}
        open={isModalOpen}
        centered={true}
        okType="danger"
        okText={t('settings.vpn.okBtn')}
        cancelText={t('settings.vpn.cancelBtn')}
        onOk={uninstall}
        onCancel={() => setIsModalOpen(false)}
        confirmLoading={isLoading}
      >
        <div className="py-5">
          <p className="text-base">{t('settings.vpn.uninstallDesc', { name: vpn.title })}</p>
          <ErrorDetail message={errMsg} />
        </div>
      </Modal>
    </>
  );
};
