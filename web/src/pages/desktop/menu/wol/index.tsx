import {
  ButtonHTMLAttributes,
  ChangeEvent,
  forwardRef,
  ReactNode,
  useEffect,
  useRef,
  useState
} from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, Divider, Input, List, message, Popconfirm, Tooltip } from 'antd';
import type { InputRef } from 'antd';
import clsx from 'clsx';
import {
  EyeIcon,
  EyeOffIcon,
  NetworkIcon,
  Pencil,
  PlusIcon,
  SendIcon,
  Trash2Icon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { WolDevice } from '@/api/network.ts';
import {
  checkWolDevice,
  deleteWolMac,
  getWolDevices,
  getWolMacs,
  setWolMacName,
  wol
} from '@/api/network.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';
import { MenuItem } from '@/components/menu-item.tsx';

import { DeviceForm, emptyForm } from './device-form.ts';
import { DeviceFormView } from './device-form.tsx';

interface MacItem {
  name: string;
  mac: string;
  isShow: boolean;
  isName: boolean;
  isEdit: boolean;
  // From the managed list: the host address, and whether it answered the last check.
  device?: WolDevice;
  up?: boolean | null;
}

// How long the list keeps checking a host after a wake, and how often.
const awakeChecks = 24;
const awakeInterval = 5000;

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

// Wol is a row of the Tools menu, with its own popover to the side.
export const Wol = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const managed = useServerFeature('wol-devices');
  // Renaming and deleting a saved address are admin actions on the server.
  const isAdmin = account.role === 'admin';

  // The rename input lives in the popover, so this lock covers it too.
  const [isOpen, setIsOpen] = useState(false);
  useKeyboardLock('wol-popover', isOpen);

  const [input, setInput] = useState('');
  const [status, setStatus] = useState('');
  const [log, setLog] = useState('');

  const [macList, setMacList] = useState<MacItem[]>([]);
  const [form, setForm] = useState<DeviceForm | null>(null);

  // Checks that run after a wake; closed with the popover.
  const timers = useRef<Map<string, number>>(new Map());
  const alive = useRef(true);
  useEffect(() => {
    alive.current = true;
    const running = timers.current;
    return () => {
      alive.current = false;
      running.forEach((id) => window.clearTimeout(id));
      running.clear();
    };
  }, []);

  const inputRef = useRef<InputRef>(null);

  function handleOpenChange(open: boolean) {
    setIsOpen(open);
    if (open) {
      getMacs();
    } else {
      setInput('');
      setStatus('');
      setLog('');
      setForm(null);
      timers.current.forEach((id) => window.clearTimeout(id));
      timers.current.clear();
    }
  }

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    setInput(e.target.value);
  }

  function setUp(mac: string, up: boolean | null) {
    if (!alive.current) return;
    setMacList((list) => list.map((item) => (item.mac === mac ? { ...item, up } : item)));
  }

  function checkOnce(mac: string): Promise<boolean> {
    return checkWolDevice(mac)
      .then((rsp) => (rsp.code === 0 ? !!rsp.data.up : false))
      .catch(() => false);
  }

  // Checks every listed host that has an address, once.
  function checkAll(devices: WolDevice[]) {
    devices
      .filter((d) => d.ip)
      .forEach((d) => {
        setUp(d.mac, null);
        checkOnce(d.mac).then((up) => setUp(d.mac, up));
      });
  }

  // After a wake, ask again until the host answers or the checks run out.
  function watchAwake(mac: string, left = awakeChecks) {
    const previous = timers.current.get(mac);
    if (previous) window.clearTimeout(previous);

    timers.current.set(
      mac,
      window.setTimeout(async () => {
        timers.current.delete(mac);
        const up = await checkOnce(mac);
        setUp(mac, up);
        if (!up && left > 1 && alive.current) watchAwake(mac, left - 1);
      }, awakeInterval)
    );
  }

  function getDevices() {
    getWolDevices()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        const devices = rsp.data.devices as WolDevice[];
        setMacList(
          devices.map((device) => ({
            name: device.name,
            mac: device.mac,
            isShow: device.name === '',
            isName: device.name !== '',
            isEdit: false,
            device,
            up: device.ip ? null : undefined
          }))
        );
        checkAll(devices);
      })
      .catch(() => {});
  }

  function getMacs() {
    if (managed) {
      getDevices();
      return;
    }

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
        const target = macList.find((item) => item.mac === value.toUpperCase());
        if (managed && target?.device?.ip) {
          setUp(target.mac, false);
          watchAwake(target.mac);
        }
      })
      .catch(() => {
        setStatus('failed');
        setLog(t('wol.requestFailed'));
      });
  }

  function openForm(item?: MacItem) {
    if (!item) {
      setForm({ ...emptyForm });
      return;
    }
    const d = item.device;
    setForm({
      ...emptyForm,
      oldMac: item.mac,
      name: item.name,
      mac: item.mac,
      broadcast: d?.broadcast ?? '',
      port: d?.port ?? 0,
      hasSecureOn: d?.hasSecureOn ?? false,
      ip: d?.ip ?? ''
    });
  }

  const content = form ? (
    <DeviceFormView
      initial={form}
      onCancel={() => setForm(null)}
      onSaved={() => {
        setForm(null);
        getMacs();
      }}
    />
  ) : (
    <div className="min-w-[300px]">
      <div className="flex items-center justify-between px-1">
        <span className="text-base font-bold text-neutral-300">{t('wol.title')}</span>
        {managed && isAdmin && (
          <IconButton
            label={t('wol.add')}
            className="text-neutral-400 hover:bg-neutral-700"
            onClick={() => openForm()}
          >
            <PlusIcon size={15} />
          </IconButton>
        )}
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
                    defaultValue={item.name}
                    onPressEnter={(e) => setMacName(e, item.mac)}
                  />
                ) : item.isShow ? (
                  item.mac
                ) : (
                  item.name
                )}
              </div>

              {item.device?.ip && (
                <span
                  className={clsx(
                    'text-xs',
                    item.up
                      ? 'text-green-500'
                      : item.up === null
                        ? 'text-neutral-500'
                        : 'text-red-400'
                  )}
                >
                  {item.up ? t('wol.up') : item.up === null ? t('wol.checking') : t('wol.down')}
                </span>
              )}

              <div className="flex items-center space-x-1">
                {item.isName && (
                  <IconButton
                    label={item.isShow ? t('wol.showName') : t('wol.showMac')}
                    className="text-neutral-400 hover:bg-neutral-700/80"
                    onClick={() => toggleShow(item.mac)}
                  >
                    {item.isShow ? <EyeOffIcon size={15} /> : <EyeIcon size={15} />}
                  </IconButton>
                )}
                {isAdmin && (
                  <IconButton
                    label={t('wol.rename')}
                    className="text-neutral-400 hover:bg-neutral-700"
                    onClick={() => (managed ? openForm(item) : editMac(item.mac, item.isEdit))}
                  >
                    <Pencil size={15} />
                  </IconButton>
                )}
                <IconButton
                  label={t('wol.wake')}
                  className="text-green-500 hover:bg-neutral-700/80"
                  onClick={() => wake(item.mac)}
                >
                  <SendIcon size={15} />
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
                      <Trash2Icon size={15} />
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
      asRow
      onOpenChange={handleOpenChange}
    />
  );
};
