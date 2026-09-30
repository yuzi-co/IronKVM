import type { CollapseProps } from 'antd';
import { Collapse, Modal } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

// TipsButton sits in the Media menu's title row. The dialog it opens is
// TipsModal, rendered outside the menu's popover: the menu closes first, so the
// dialog is not drawn under or over a dropdown left open behind it.
export const TipsButton = ({ onClick }: { onClick: () => void }) => {
  const { t } = useTranslation();

  return (
    <button
      type="button"
      aria-label={t('image.tips.title')}
      className="flex cursor-pointer items-center space-x-1 p-0 text-neutral-500 hover:text-blue-500"
      onClick={onClick}
    >
      <CircleHelpIcon size={14} />
    </button>
  );
};

type TipsModalProps = {
  open: boolean;
  onClose: () => void;
};

export const TipsModal = ({ open, onClose }: TipsModalProps) => {
  const { t } = useTranslation();
  useKeyboardLock('image-tips', open);

  const items: CollapseProps['items'] = [
    {
      key: '1',
      label: 'USB',
      children: (
        <ul className="list-decimal">
          <li>{t('image.tips.usb1')}</li>
          <li>{t('image.tips.usb2')}</li>
          <li>{t('image.tips.usb3')}</li>
        </ul>
      )
    },
    {
      key: '2',
      label: 'SCP',
      children: (
        <ul className="list-decimal">
          <li>{t('image.tips.scp1')}</li>
          <li>{t('image.tips.scp2')}</li>
          <li>{t('image.tips.scp3')}</li>
        </ul>
      )
    },
    {
      key: '3',
      label: t('image.tips.tfCard'),
      children: (
        <>
          <p className="pl-5 text-sm text-neutral-400">{t('image.tips.tf1')}</p>
          <ul className="list-decimal">
            <li>{t('image.tips.tf2')}</li>
            <li>{t('image.tips.tf3')}</li>
            <li>{t('image.tips.tf4')}</li>
            <li>{t('image.tips.tf5')}</li>
          </ul>
        </>
      )
    }
  ];

  return (
    <Modal
      title={t('image.tips.title')}
      open={open}
      width={520}
      footer={null}
      centered
      onCancel={onClose}
    >
      <Collapse
        accordion
        items={items}
        bordered={false}
        defaultActiveKey={['1']}
        style={{ backgroundColor: 'transparent' }}
      />
    </Modal>
  );
};
