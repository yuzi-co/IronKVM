import { useEffect, useState } from 'react';
import { Divider, InputNumber, Switch } from 'antd';
import { useAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { isHdmiEnabledAtom } from '@/jotai/screen.ts';

import { Section } from './section.tsx';

export const Hdmi = () => {
  const { t } = useTranslation();

  const [isHdmiEnabled, setIsHdmiEnabled] = useAtom(isHdmiEnabledAtom);

  const [isPcie, setIsPcie] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [idleTimeout, setIdleTimeout] = useState(0);
  const [idleTimeoutInput, setIdleTimeoutInput] = useState<number | null>(0);
  const [isIdleTimeoutLoading, setIsIdleTimeoutLoading] = useState(false);

  useEffect(() => {
    async function getHardware() {
      try {
        const rsp = await api.getHardware();
        if (rsp.code === 0) setIsPcie(rsp.data?.version === 'PCIE');
      } catch {
        // Not a PCIe board as far as this page knows; the section stays hidden.
      }
    }

    async function getHdmiState() {
      try {
        const rsp = await api.getHdmiState();
        if (rsp.code === 0) {
          setIsHdmiEnabled(rsp.data.enabled);
          const timeout = rsp.data.idleTimeout ?? 0;
          setIdleTimeout(timeout);
          setIdleTimeoutInput(timeout);
        }
      } catch (err) {
        showFailure(err);
      } finally {
        setIsLoading(false);
      }
    }

    getHardware();
    getHdmiState();
  }, [setIsHdmiEnabled]);

  function updateIdleTimeout() {
    if (
      isIdleTimeoutLoading ||
      idleTimeoutInput === null ||
      idleTimeoutInput < 0 ||
      idleTimeoutInput === idleTimeout
    ) {
      return;
    }

    setIsIdleTimeoutLoading(true);
    api
      .setHdmiIdleTimeout(idleTimeoutInput)
      .then((rsp) => {
        if (!showResult(rsp, { success: t('feedback.saved') })) {
          setIdleTimeoutInput(idleTimeout);
          return;
        }

        setIdleTimeout(idleTimeoutInput);
      })
      .catch((err) => {
        setIdleTimeoutInput(idleTimeout);
        showFailure(err);
      })
      .finally(() => {
        setIsIdleTimeoutLoading(false);
      });
  }

  async function setHdmiState() {
    if (isLoading) return;
    setIsLoading(true);

    const enabled = !isHdmiEnabled;

    // The switch spins until the answer, and a thrown request must stop it
    // as surely as a refusal does.
    let keepSpinning = false;
    try {
      const rsp = await api.setHdmiState(enabled);
      if (!showResult(rsp)) return;

      keepSpinning = true;
      setTimeout(() => {
        setIsHdmiEnabled(enabled);
        setIsLoading(false);
      }, 1000);
    } catch (err) {
      showFailure(err);
    } finally {
      if (!keepSpinning) setIsLoading(false);
    }
  }

  if (!isPcie) return null;

  return (
    <>
      <Section title={t('settings.device.sections.video')}>
        <div className="flex flex-col space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex flex-col space-y-1">
              <span>HDMI</span>

              <span className="text-xs text-neutral-500">
                {t('settings.device.hdmi.description')}
              </span>
            </div>

            <Switch checked={isHdmiEnabled} loading={isLoading} onChange={setHdmiState} />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex flex-col space-y-1">
              <span>{t('settings.device.hdmi.idleTimeoutTitle')}</span>
              <span className="text-xs text-neutral-500">
                {t('settings.device.hdmi.idleTimeoutDescription')}
              </span>
            </div>

            <InputNumber
              style={{ width: 150 }}
              min={0}
              max={10080}
              precision={0}
              value={idleTimeoutInput}
              addonAfter={t('settings.device.hdmi.minutes')}
              disabled={isIdleTimeoutLoading}
              onChange={setIdleTimeoutInput}
              onBlur={updateIdleTimeout}
              onPressEnter={updateIdleTimeout}
            />
          </div>
        </div>
      </Section>
      <Divider className="opacity-50" style={{ margin: 0 }} />
    </>
  );
};
