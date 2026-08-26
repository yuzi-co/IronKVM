import { useEffect, useState } from 'react';
import { Popconfirm, Select, Tooltip } from 'antd';
import { CircleAlertIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';

type CpuFreqState = {
  running: number;
  measured: boolean;
  target: number;
  temperature: number;
  options: number[];
  rebootRequired: boolean;
};

export const CpuFreq = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [isRebooting, setIsRebooting] = useState(false);
  const [state, setState] = useState<CpuFreqState | null>(null);

  useEffect(() => {
    getCpuFreq();
  }, []);

  function getCpuFreq() {
    setIsLoading(true);

    api
      .getCpuFreq()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }
        setState(rsp.data);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  function update(value: number) {
    if (isLoading) return;
    setIsLoading(true);

    api
      .setCpuFreq(value)
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
        }
      })
      .finally(() => {
        // Read the state back either way: the change only takes effect on the
        // next boot, so what matters now is the target the server recorded.
        getCpuFreq();
      });
  }

  function reboot() {
    if (isRebooting) return;
    setIsRebooting(true);

    const timeoutId = setTimeout(() => {
      window.location.reload();
    }, 30000);

    api
      .reboot()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          setIsRebooting(false);
          clearTimeout(timeoutId);
        }
      })
      .catch((err) => {
        console.log(err);
        setIsRebooting(false);
        clearTimeout(timeoutId);
      });
  }

  const options = (state?.options ?? []).map((mhz) => ({ value: mhz, label: `${mhz} MHz` }));

  // status is the compact line under the description. It reports the clock the
  // core runs now and the temperature, and turns amber with a reboot prompt
  // when the running clock no longer matches the recorded target.
  function statusText(): string {
    if (!state) return '';

    const parts: string[] = [];
    if (state.measured && state.running > 0) {
      parts.push(t('settings.device.cpuFreq.running', { mhz: state.running }));
    }
    if (state.temperature > 0) {
      parts.push(`${state.temperature.toFixed(1)}°C`);
    }
    return parts.join(' · ');
  }

  const text = statusText();

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>{t('settings.device.cpuFreq.title')}</span>

          <Tooltip
            title={t('settings.device.cpuFreq.tip')}
            className="cursor-pointer"
            placement="right"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleAlertIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>

        <span className="text-xs text-neutral-500">{t('settings.device.cpuFreq.description')}</span>

        {state && text && (
          <>
            {state.rebootRequired ? (
              <Popconfirm
                placement="right"
                title={t('settings.device.cpuFreq.rebootConfirm', { mhz: state.target })}
                okText={t('settings.device.okBtn')}
                cancelText={t('settings.device.cancelBtn')}
                okButtonProps={{ loading: isRebooting }}
                onConfirm={reboot}
              >
                <span className="w-fit cursor-pointer text-xs text-amber-500">
                  {text} · {t('settings.device.cpuFreq.rebootToApply')}
                </span>
              </Popconfirm>
            ) : (
              <span className="w-fit cursor-default text-xs text-neutral-400">{text}</span>
            )}
          </>
        )}
      </div>

      <Select
        style={{ width: 150 }}
        value={state?.target}
        options={options}
        loading={isLoading}
        disabled={!state || options.length === 0}
        onChange={update}
      />
    </div>
  );
};
