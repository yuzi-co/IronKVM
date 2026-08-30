import { useEffect, useState } from 'react';
import { Select } from 'antd';
import { ScreenShareOff } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';

// The drive current for the panel, as SSD1306 command 0x81 takes it.
//
// 207 is what the firmware writes at start, and the levels below it are there
// because this panel burns in: it dims where it has been lit, and a lower drive
// slows that down everywhere. The lowest offered is 64 rather than the 16 the
// server allows, because anything under that is hard to read in a lit room and
// this control is the only way back.
const BRIGHTNESS_LEVELS = [64, 96, 128, 160, 207, 255];

export const Oled = () => {
  const { t } = useTranslation();

  const [isOLEDExist, setIsOLEDExist] = useState(false);
  const [isSleepLoading, setIsSleepLoading] = useState(false);
  const [isBrightnessLoading, setIsBrightnessLoading] = useState(false);
  const [sleep, setSleep] = useState('');
  const [brightness, setBrightness] = useState('');
  const [isBrightnessSupported, setIsBrightnessSupported] = useState(false);

  useEffect(() => {
    api.getOLED().then((rsp) => {
      if (rsp.code !== 0) {
        console.log(rsp.msg);
        return;
      }

      if (!rsp.data.exist) {
        return;
      }

      setIsOLEDExist(true);
      setSleep(rsp.data.sleep.toString());
      setBrightness(rsp.data.brightness.toString());
      setIsBrightnessSupported(rsp.data.brightnessSupported);
    });
  }, []);

  const sleepOptions = [0, 15, 30, 60, 180, 300, 600, 1800, 3600].map((duration) => ({
    value: duration.toString(),
    label: t(`settings.device.oled.${duration}`)
  }));

  const brightnessOptions = BRIGHTNESS_LEVELS.map((level) => ({
    value: level.toString(),
    label: t(`settings.device.oled.brightnessLevels.${level}`)
  }));

  function updateSleep(value: string) {
    if (isSleepLoading) return;
    setIsSleepLoading(true);

    api
      .setOLED({ sleep: parseInt(value) })
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        setSleep(value);
      })
      .finally(() => {
        setIsSleepLoading(false);
      });
  }

  function updateBrightness(value: string) {
    if (isBrightnessLoading) return;
    setIsBrightnessLoading(true);

    api
      .setOLED({ brightness: parseInt(value) })
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        setBrightness(value);
      })
      .finally(() => {
        setIsBrightnessLoading(false);
      });
  }

  return (
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.device.oled.title')}</span>
          <span className="text-xs text-neutral-500">{t('settings.device.oled.description')}</span>
        </div>

        {isOLEDExist ? (
          <Select
            style={{ width: 150 }}
            value={sleep}
            options={sleepOptions}
            loading={isSleepLoading}
            onChange={updateSleep}
          />
        ) : (
          <span className="text-neutral-500">
            <ScreenShareOff size={16} />
          </span>
        )}
      </div>

      {/* The control appears only when the running firmware acts on it. A
          release carries Sipeed's kvm_system, which reads the sleep setting and
          ignores this one, and a control that changes nothing is worse than an
          absent one. */}
      {isOLEDExist && isBrightnessSupported && (
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1">
            <span>{t('settings.device.oled.brightness')}</span>
            <span className="text-xs text-neutral-500">
              {t('settings.device.oled.brightnessDescription')}
            </span>
          </div>

          <Select
            style={{ width: 150 }}
            value={brightness}
            options={brightnessOptions}
            loading={isBrightnessLoading}
            onChange={updateBrightness}
          />
        </div>
      )}
    </>
  );
};
