import { ChangeEvent, useState } from 'react';
import { Button, Input, InputNumber, Modal, Radio, RadioChangeEvent, Select } from 'antd';
import { SquareTerminalIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { getSerialSettings, setSerialSettings } from '@/lib/localstorage.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';
import { isValidSerialPort, validatePicocomParameters } from '@/pages/terminal/validater.ts';

export const SerialPort = () => {
  const { t } = useTranslation();

  // The dialog opens with the settings last used, read once per mount.
  const [saved] = useState(() => getSerialSettings() ?? {});

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [port, setPort] = useState(saved.port ?? '');
  const [baudrate, setBaudrate] = useState(saved.baudrate ?? 115200);
  const [parity, setParity] = useState<string>(saved.parity ?? 'none');
  const [flowControl, setFlowControl] = useState<string>(saved.flowControl ?? 'none');
  const [dataBits, setDataBits] = useState(saved.dataBits ?? 8);
  const [stopBits, setStopBits] = useState(saved.stopBits ?? 1);
  const [error, setError] = useState<'' | 'invalidPort' | 'invalidBaud'>('');

  useKeyboardLock('serial-port-modal', isModalOpen);

  function openModal() {
    setError('');
    setIsModalOpen(true);
  }

  function closeModal() {
    setIsModalOpen(false);
  }

  function handleInputChange(e: ChangeEvent<HTMLInputElement>) {
    setPort(e.target.value);
    setError('');
  }

  function handleRadioChange(e: RadioChangeEvent) {
    setPort(e.target.value);
    setError('');
  }

  function onInputNumberChange(value: number | null) {
    if (value === null) return;
    setBaudrate(value);
    setError('');
  }

  // The terminal page runs the same check and, when it fails, opens a plain
  // shell. Checking here says what is wrong instead.
  function submit() {
    const trimmed = port.trim();
    if (!isValidSerialPort(trimmed)) {
      setError('invalidPort');
      return;
    }

    const valid = validatePicocomParameters({
      port: trimmed,
      baud: String(baudrate),
      parity,
      flowControl,
      dataBits: String(dataBits),
      stopBits: String(stopBits)
    });
    if (!valid) {
      setError('invalidBaud');
      return;
    }

    setSerialSettings({ port: trimmed, baudrate, parity, flowControl, dataBits, stopBits });
    setIsModalOpen(false);

    const query = new URLSearchParams({
      port: trimmed,
      baud: String(baudrate),
      parity,
      flowControl,
      dataBits: String(dataBits),
      stopBits: String(stopBits)
    });
    window.open(`/#terminal?${query.toString()}`, '_blank');
  }

  return (
    <>
      <button
        type="button"
        className="flex h-[28px] w-full cursor-pointer items-center space-x-1 rounded p-0 px-2 py-1 text-left select-none hover:bg-neutral-700/70"
        onClick={openModal}
      >
        <SquareTerminalIcon size={14} />
        <span>{t('terminal.serial')}</span>
      </button>

      <Modal open={isModalOpen} title={t('terminal.serial')} footer={null} onCancel={closeModal}>
        <div className="mt-10 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">
            {t('terminal.serialPort')}
          </div>
          <div className="w-1/2">
            <Input
              value={port}
              status={error === 'invalidPort' ? 'error' : undefined}
              placeholder={t('terminal.serialPortPlaceholder')}
              onChange={handleInputChange}
            />
          </div>
        </div>
        <div className="mt-3 pl-[100px]">
          <Radio.Group size="large" value={port.trim()} onChange={handleRadioChange}>
            <Radio value="/dev/ttyS1">
              <code>/dev/ttyS1</code>
            </Radio>
            <Radio value="/dev/ttyS2">
              <code>/dev/ttyS2</code>
            </Radio>
          </Radio.Group>
        </div>

        <div className="mt-7 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">{t('terminal.baudrate')}</div>
          <InputNumber controls={false} min={1} value={baudrate} onChange={onInputNumberChange} />
        </div>

        <div className="mt-7 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">{t('terminal.parity')}</div>
          <div className="w-1/2">
            <Select
              defaultValue="none"
              value={parity}
              onChange={setParity}
              options={[
                { label: t('terminal.parityNone'), value: 'none' },
                { label: t('terminal.parityEven'), value: 'even' },
                { label: t('terminal.parityOdd'), value: 'odd' }
              ]}
            />
          </div>
        </div>

        <div className="mt-7 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">
            {t('terminal.flowControl')}
          </div>
          <div className="w-1/2">
            <Select
              defaultValue="none"
              value={flowControl}
              onChange={setFlowControl}
              options={[
                { label: t('terminal.flowControlNone'), value: 'none' },
                { label: t('terminal.flowControlSoft'), value: 'soft' },
                { label: t('terminal.flowControlHard'), value: 'hard' }
              ]}
            />
          </div>
        </div>

        <div className="mt-7 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">{t('terminal.dataBits')}</div>
          <div className="w-1/2">
            <Select
              defaultValue={8}
              value={dataBits}
              onChange={setDataBits}
              options={[
                { label: '5', value: 5 },
                { label: '6', value: 6 },
                { label: '7', value: 7 },
                { label: '8', value: 8 }
              ]}
            />
          </div>
        </div>

        <div className="mt-7 flex items-center space-x-[20px]">
          <div className="flex w-[80px] justify-end text-neutral-400">{t('terminal.stopBits')}</div>
          <div className="w-1/2">
            <Select
              defaultValue={1}
              value={stopBits}
              onChange={setStopBits}
              options={[
                { label: '1', value: 1 },
                { label: '2', value: 2 }
              ]}
            />
          </div>
        </div>

        {error && (
          <div className="mt-7 text-center text-sm text-red-500">{t(`terminal.${error}`)}</div>
        )}

        <div className="mt-12 mb-3 flex justify-center">
          <Button type="primary" onClick={submit}>
            {t('terminal.confirm')}
          </Button>
        </div>
      </Modal>
    </>
  );
};
