import { useEffect, useState } from 'react';
import { Button, Card, Modal, Spin } from 'antd';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/script';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

type RunProps = {
  script: string;
  setIsRunning: (isRunning: boolean) => void;
};

export const Run = ({ script, setIsRunning }: RunProps) => {
  const { t } = useTranslation();
  // The request is sent on mount, so 'running' is the state the first paint
  // should already show.
  const [state, setState] = useState('running');
  const [log, setLog] = useState('');

  // The dialog is open for as long as it is mounted.
  useKeyboardLock('script-run', true);

  useEffect(() => {
    api
      .runScript(script, 'foreground')
      .then((rsp) => {
        if (rsp.code !== 0) {
          // A script that exits with an error still printed why.
          const output = rsp.data?.log;
          setLog(output ? `${rsp.msg}\n\n${output}` : rsp.msg);
          setState('failed');
          return;
        }

        setState('success');
        setLog(rsp.data.log);
      })
      .catch((err) => {
        // The sentence is chosen at render, so the effect does not depend on
        // t. A language change must not run the script a second time.
        setState(err?.code === 'ECONNABORTED' ? 'timedOut' : 'unreachable');
      });
  }, [script]);

  return (
    <Modal
      title={script}
      width={600}
      closable={false}
      footer={null}
      style={{ top: 60 }}
      open={true}
    >
      {state === 'running' ? (
        <div className="flex h-[300px] flex-col items-center justify-center space-y-6">
          <Spin indicator={<LoaderCircleIcon size={32} className="animate-spin" />} size="large" />
          <span className="text-xs text-neutral-500">
            {t('script.waitLimit', { minutes: api.FOREGROUND_TIMEOUT_MINUTES })}
          </span>
        </div>
      ) : (
        <Card className="h-[600px] overflow-auto font-mono whitespace-pre-line">
          {state === 'unreachable'
            ? t('script.runFailed')
            : state === 'timedOut'
              ? t('script.timedOut', { minutes: api.FOREGROUND_TIMEOUT_MINUTES })
              : log}
        </Card>
      )}

      <div className="mt-5 flex justify-center">
        <Button type="primary" onClick={() => setIsRunning(false)}>
          {t('script.close')}
        </Button>
      </div>
    </Modal>
  );
};
