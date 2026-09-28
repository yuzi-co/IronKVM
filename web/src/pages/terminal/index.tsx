import { useEffect } from 'react';
import { AttachAddon } from '@xterm/addon-attach';
import { FitAddon } from '@xterm/addon-fit';
import { Terminal as XtermTerminal } from '@xterm/xterm';
import i18n from 'i18next';
import { useTranslation } from 'react-i18next';

import '@xterm/xterm/css/xterm.css';

import { notifyAuthExpired } from '@/lib/auth-events.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { Head } from '@/components/head.tsx';

import { validatePicocomParameters } from './validater.ts';

export const Terminal = () => {
  const { t } = useTranslation();

  useEffect(() => {
    const terminalEle = document.getElementById('terminal');
    if (!terminalEle) return;

    const terminal = new XtermTerminal({
      cursorBlink: true
    });

    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(terminalEle);
    fitAddon.fit();

    const url = `${getBaseUrl('ws')}/api/vm/terminal`;
    let ws: WebSocket;
    let attachAddon: AttachAddon | null = null;
    let isPicocomRunning = false;
    let isDisconnected = false;
    let isDisposed = false;

    const sendSize = () => {
      if (ws.readyState !== WebSocket.OPEN) return;
      const windowSize = { rows: terminal.rows, cols: terminal.cols };
      const blob = new Blob([JSON.stringify(windowSize)], { type: 'application/json' });
      ws.send(blob);
    };

    const runPicocom = () => {
      const urls = window.location.href.split('?');
      if (urls.length < 2) return;

      const searchParams = new URLSearchParams(urls[1]);
      const port = searchParams.get('port');
      const baud = searchParams.get('baud');
      const parity = searchParams.get('parity');
      const flowControl = searchParams.get('flowControl');
      const dataBits = searchParams.get('dataBits');
      const stopBits = searchParams.get('stopBits');
      if (!port || !baud) return;

      // The shell is already open; say why it is not the serial port.
      if (!validatePicocomParameters({ port, baud, parity, flowControl, dataBits, stopBits })) {
        terminal.writeln(`\x1b[31m[${i18n.t('terminal.invalidSettings')}]\x1b[0m`);
        return;
      }

      ws.send(
        `picocom ${port} --baud ${baud} --parity ${parity} --flow ${flowControl} --databits ${dataBits} --stopbits ${stopBits}\r`
      );

      isPicocomRunning = true;
    };

    // connect opens the shell socket. A dropped socket leaves the page with a
    // note and waits for Enter, rather than a terminal that looks alive but
    // swallows every key.
    const connect = () => {
      isDisconnected = false;
      ws = new WebSocket(url);

      ws.addEventListener('close', (event) => {
        attachAddon?.dispose();
        attachAddon = null;
        isPicocomRunning = false;

        if (isDisposed) return;

        if (event.code === 4401) {
          notifyAuthExpired();
          return;
        }

        isDisconnected = true;
        terminal.writeln('');
        terminal.writeln(`\x1b[33m[${i18n.t('terminal.disconnected')}]\x1b[0m`);
      });

      ws.onopen = () => {
        attachAddon = new AttachAddon(ws);
        terminal.loadAddon(attachAddon);

        sendSize();
        setTimeout(runPicocom, 300);
      };
    };

    const onData = terminal.onData((data) => {
      if (isDisconnected && (data === '\r' || data === '\n')) {
        terminal.writeln('');
        connect();
      }
    });

    connect();

    const exitPicocom = () => {
      if (ws.readyState === WebSocket.OPEN && isPicocomRunning) {
        ws.send('\x01\x18');
        isPicocomRunning = false;
      }
    };

    const resizeScreen = () => {
      fitAddon.fit();
      sendSize();
    };

    const cleanupConnection = () => {
      exitPicocom();
      const socket = ws;
      setTimeout(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.close();
        }
      }, 100);
    };

    const handleBeforeUnload = () => {
      isDisposed = true;
      cleanupConnection();
    };

    window.addEventListener('resize', resizeScreen, false);
    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      isDisposed = true;
      onData.dispose();
      terminal.dispose();
      cleanupConnection();

      window.removeEventListener('resize', resizeScreen, false);
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);

  return (
    <>
      <Head title={t('head.terminal')} />

      <div className="h-full w-full overflow-hidden">
        <div id="terminal" className="h-full p-2"></div>
      </div>
    </>
  );
};
