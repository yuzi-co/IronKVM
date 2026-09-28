import { useEffect, useRef, useState } from 'react';
import { LoadingOutlined, RocketOutlined, SmileOutlined } from '@ant-design/icons';
import { Button, Divider, Popconfirm, Result, Spin } from 'antd';
import { useTranslation } from 'react-i18next';
import semver from 'semver';

import * as api from '@/api/application.ts';
import {
  reloadAfterRestart,
  SERVER_RESTART_DOWN_MS,
  SERVER_RESTART_UP_MS
} from '@/lib/wait-server.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { CustomServer } from './custom-server.tsx';
import { Offline } from './offline.tsx';
import { Preview } from './preview.tsx';

type UpdateProps = {
  setIsLocked: (isClosable: boolean) => void;
};

export const Update = ({ setIsLocked }: UpdateProps) => {
  const { t } = useTranslation();

  const [status, setStatus] = useState('');
  const [currentVersion, setCurrentVersion] = useState('');
  const [latestVersion, setLatestVersion] = useState('');
  const [errMsg, setErrMsg] = useState('');
  const [isCustomServerEnabled, setIsCustomServerEnabled] = useState(false);
  const [isCustomServerPending, setIsCustomServerPending] = useState(false);
  const versionRequestRef = useRef(0);

  const checkForUpdates = useStableCallback(() => {
    const requestId = ++versionRequestRef.current;
    setStatus('loading');

    api
      .getVersion()
      .then((rsp: any) => {
        if (requestId !== versionRequestRef.current) return;
        if (rsp.code !== 0 || !rsp.data) {
          setStatus('failed');
          setErrMsg(t('settings.update.queryFailed'));
          return;
        }

        setCurrentVersion(rsp.data.current);

        if (rsp.data?.latest) {
          setLatestVersion(rsp.data.latest);
          const isLatest = semver.gte(rsp.data.current, rsp.data.latest);
          setStatus(isLatest ? 'latest' : 'outdated');
        } else {
          setStatus('latest');
        }
      })
      .catch(() => {
        if (requestId !== versionRequestRef.current) return;
        setStatus('failed');
        setErrMsg(t('settings.update.queryFailed'));
      });
  });

  useEffect(() => {
    checkForUpdates();
  }, [checkForUpdates]);

  function update() {
    if (status !== 'outdated' || isCustomServerPending) return;

    setIsLocked(true);
    setStatus('updating');

    // The server answers once the new version is installed, and restarts
    // itself a second later. A failure keeps its message on screen: reloading
    // would only hide it.
    api
      .update()
      .then((rsp: any) => {
        if (rsp.code !== 0) {
          fail(rsp.msg);
          return;
        }
        reloadAfterRestart(SERVER_RESTART_DOWN_MS, SERVER_RESTART_UP_MS);
      })
      .catch((err: any) => fail(err?.message));
  }

  function fail(detail?: string) {
    setIsLocked(false);
    setStatus('failed');
    const failed = t('settings.update.updateFailed');
    setErrMsg(detail ? `${failed} (${detail})` : failed);
  }

  return (
    <>
      <div className="text-base">{t('settings.update.title')}</div>
      <Divider className="opacity-50" />

      <Preview
        checkForUpdates={checkForUpdates}
        disabled={isCustomServerEnabled || isCustomServerPending}
      />
      <CustomServer
        checkForUpdates={checkForUpdates}
        onEnabledChange={setIsCustomServerEnabled}
        onPendingChange={setIsCustomServerPending}
      />
      <Offline
        status={status}
        setStatus={setStatus}
        setIsLocked={setIsLocked}
        setErrMsg={setErrMsg}
      />
      <Divider className="opacity-50" />

      <div className="flex min-h-[320px] flex-col justify-between">
        {status === 'loading' && (
          <div className="flex justify-center pt-24">
            <Spin indicator={<LoadingOutlined spin />} size="large" />
          </div>
        )}

        {status === 'updating' && (
          <div className="flex flex-col items-center justify-center space-y-10 pt-24 pb-10">
            <Spin size="large" />
            <span className="text-neutral-500">{t('settings.update.updating')}</span>
          </div>
        )}

        {status === 'latest' && (
          <Result
            status="success"
            icon={<SmileOutlined />}
            title={currentVersion}
            subTitle={t('settings.update.isLatest')}
            extra={[
              <Button key="confirm" onClick={checkForUpdates}>
                {t('settings.update.title')}
              </Button>
            ]}
          />
        )}

        {status === 'outdated' && (
          <Result
            status="warning"
            icon={<RocketOutlined />}
            title={`${currentVersion} -> ${latestVersion}`}
            subTitle={t('settings.update.available')}
            extra={[
              <Popconfirm
                key="confirm"
                placement="bottom"
                title={t('settings.update.updateTo', { version: latestVersion })}
                description={
                  <div className="max-w-[320px]">{t('settings.update.updateConfirmDesc')}</div>
                }
                okText={t('settings.update.confirm')}
                cancelText={t('settings.update.cancel')}
                disabled={isCustomServerPending}
                onConfirm={update}
              >
                <Button type="primary" disabled={isCustomServerPending}>
                  {t('settings.update.updateTo', { version: latestVersion })}
                </Button>
              </Popconfirm>
            ]}
          />
        )}

        {status === 'failed' && <Result subTitle={errMsg} />}

        <div className="flex justify-center">
          <Button
            type="link"
            size="small"
            href="https://github.com/yuzi-co/IronKVM/releases"
            target="_blank"
          >
            {t('settings.update.releaseNotes')}
          </Button>
        </div>
      </div>
    </>
  );
};
