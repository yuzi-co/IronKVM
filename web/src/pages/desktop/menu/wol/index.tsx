import { ButtonHTMLAttributes, ChangeEvent, forwardRef, ReactNode, useRef, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, Divider, Input, List, message, Popconfirm, Tooltip } from 'antd';
import type { InputRef } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { Eye, EyeClosed, NetworkIcon, Pencil, SendIcon, Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { deleteWolMac, getWolMacs, setWolMacName, wol } from '@/api/network.ts';
import { keyboardLockAtom } from '@/jotai/keyboard.ts';
import { MenuItem } from '@/components/menu-item.tsx';

interface MacItem {
  name: string;
  mac: string;
  isShow: boolean;
  isName: boolean;
  isEdit: boolean;
}

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string;
  children: ReactNode;
};

// IconButton is a row action with its name as the tooltip and the accessible
// label. It forwards its ref and the other props, so a Popconfirm can wrap it.
const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ label, className, children, ...rest }, ref) => (
    <Tooltip title={label} mouseEnterDelay={0.6}>
      <button
        ref={ref}
        type="button"
        aria-label={label}
        className={clsx(
          'flex h-[24px] w-[24px] cursor-pointer items-center justify-center rounded p-0',
          className
        )}
        {...rest}
      >
        {children}
      </button>
    </Tooltip>
  )
);
IconButton.displayName = 'IconButton';

export const Wol = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  // Renaming and deleting a saved address are admin actions on the server.
  const isAdmin = account.role === 'admin';

  const setKeyboardLock = useSetAtom(keyboardLockAtom);

  const [input, setInput] = useState('');
  const [status, setStatus] = useState('');
  const [log, setLog] = useState('');

  const [macList, setMacList] = useState<MacItem[]>([]);

  const inputRef = useRef<InputRef>(null);

  function handleOpenChange(open: boolean) {
    if (open) {
      getMacs();
      setKeyboardLock({ source: 'wol-popover', locked: true });
    } else {
      setInput('');
      setStatus('');
      setLog('');
      setKeyboardLock({ source: 'wol-popover', locked: false });
      setKeyboardLock({ source: 'wol-edit-input', locked: false });
    }
  }

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    setInput(e.target.value);
  }

  function getMacs() {
    getWolMacs()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        const isEdit = false;
        const macList = rsp.data.macs
          .map((item: string) => item.trim())
          .filter((item: string) => item !== '')
          .map((item: string) => {
            const separator = item.search(/\s/);
            const mac = separator === -1 ? item : item.slice(0, separator);
            const name = separator === -1 ? '' : item.slice(separator).trim();
            const isName = name !== '';
            const isShow = !isName;
            return { name, mac, isShow, isName, isEdit };
          });

        setMacList(macList);
      })
      // The saved list is a convenience; the address field still works.
      .catch(() => {});
  }

  function toggleShow(mac: string) {
    setMacList((list) =>
      list.map((item) => (item.mac === mac ? { ...item, isShow: !item.isShow } : item))
    );
  }

  function editMac(mac: string, isEdit: boolean) {
    setMacList((list) =>
      list.map((item) => (item.mac === mac ? { ...item, isEdit: !isEdit } : item))
    );
  }

  async function setMacName(e: React.KeyboardEvent<HTMLInputElement>, mac: string) {
    const value = e.currentTarget.value.trim();
    if (!value) return;

    try {
      const rsp = await setWolMacName(mac, value);
      if (rsp.code !== 0) {
        message.error(rsp.msg ? `${t('wol.renameFailed')}: ${rsp.msg}` : t('wol.renameFailed'));
        return;
      }
    } catch {
      message.error(t('wol.renameFailed'));
      return;
    }
    getMacs();
  }

  function deleteMac(mac: string) {
    deleteWolMac(mac)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg ? `${t('wol.deleteFailed')}: ${rsp.msg}` : t('wol.deleteFailed'));
          return;
        }
        getMacs();
      })
      .catch(() => message.error(t('wol.deleteFailed')));
  }

  function wake(mac?: string) {
    if (status === 'loading') return;

    const value = (mac ? mac : input).trim();
    if (!value) return;

    setStatus('loading');
    setLog(t('wol.sending'));

    wol(value)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setStatus('failed');
          setLog(rsp.msg);
          return;
        }

        setStatus('success');
        setLog(t('wol.sent'));
        getMacs();
        setInput('');
      })
      .catch(() => {
        setStatus('failed');
        setLog(t('wol.requestFailed'));
      });
  }

  const content = (
    <div className="min-w-[300px]">
      <div className="flex items-center justify-between px-1">
        <span className="text-base font-bold text-neutral-300">{t('wol.title')}</span>
      </div>

      <Divider style={{ margin: '10px 0 10px 0' }} />

      <div className="w-full space-y-1 py-3">
        <div className="flex items-center space-x-1">
          <Input
            ref={inputRef}
            value={input}
            placeholder={t('wol.input')}
            onChange={handleChange}
            onPressEnter={() => wake()}
          />
          <Button type="primary" onClick={() => wake()}>
            {t('wol.ok')}
          </Button>
        </div>

        {status && (
          <div
            className={clsx(
              'max-w-[300px] text-sm wrap-break-word',
              status === 'failed' ? 'text-red-500' : 'text-green-500'
            )}
          >
            {log}
          </div>
        )}
      </div>

      {macList.length > 0 && (
        <List
          itemLayout="horizontal"
          dataSource={macList}
          renderItem={(item) => (
            <List.Item className="flex w-full items-center justify-between">
              <div className="h-[24px] max-w-[200px]">
                {item.isEdit ? (
                  <Input
                    placeholder={item.mac}
                    onFocus={() => setKeyboardLock({ source: 'wol-edit-input', locked: true })}
                    onBlur={() => setKeyboardLock({ source: 'wol-edit-input', locked: false })}
                    defaultValue={item.name}
                    onPressEnter={(e) => setMacName(e, item.mac)}
                  />
                ) : item.isShow ? (
                  item.mac
                ) : (
                  item.name
                )}
              </div>

              <div className="flex items-center space-x-1">
                {item.isName && (
                  <IconButton
                    label={item.isShow ? t('wol.showName') : t('wol.showMac')}
                    className="text-neutral-400 hover:bg-neutral-700/80"
                    onClick={() => toggleShow(item.mac)}
                  >
                    {item.isShow ? <EyeClosed size={16} /> : <Eye size={16} />}
                  </IconButton>
                )}
                {isAdmin && (
                  <IconButton
                    label={t('wol.rename')}
                    className="text-neutral-400 hover:bg-neutral-700"
                    onClick={() => editMac(item.mac, item.isEdit)}
                  >
                    <Pencil size={16} />
                  </IconButton>
                )}
                <IconButton
                  label={t('wol.wake')}
                  className="text-green-500 hover:bg-neutral-700/80"
                  onClick={() => wake(item.mac)}
                >
                  <SendIcon size={16} />
                </IconButton>
                {isAdmin && (
                  <Popconfirm
                    title={t('wol.deleteConfirm')}
                    description={item.isName ? `${item.name} (${item.mac})` : item.mac}
                    okText={t('wol.yes')}
                    cancelText={t('wol.no')}
                    okButtonProps={{ danger: true }}
                    onConfirm={() => deleteMac(item.mac)}
                  >
                    <IconButton
                      label={t('wol.delete')}
                      className="text-red-500 hover:bg-neutral-700"
                    >
                      <Trash2Icon size={16} />
                    </IconButton>
                  </Popconfirm>
                )}
              </div>
            </List.Item>
          )}
        />
      )}
    </div>
  );

  return (
    <MenuItem
      title={t('wol.title')}
      icon={<NetworkIcon size={18} />}
      content={content}
      onOpenChange={handleOpenChange}
    />
  );
};
