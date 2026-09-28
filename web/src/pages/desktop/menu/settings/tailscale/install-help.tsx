import { Card } from 'antd';
import { useTranslation } from 'react-i18next';

// The add-on directory and the links file are the ones the install from the
// page writes (server/service/extensions/addon): S04addons reads them at boot
// and links the binaries into /usr/bin and /usr/sbin, so a hand install
// survives an image update the same way.
const addonDir = '/data/ironkvm/addons/tailscale/';
const links = '/usr/bin/tailscale tailscale\n/usr/sbin/tailscaled tailscaled';

// InstallHelp is the manual install, shown when the install from the page
// fails.
export const InstallHelp = () => {
  const { t } = useTranslation();

  return (
    <Card styles={{ body: { padding: 0 } }}>
      <p className="px-4 pt-3 text-left text-sm text-neutral-400">
        {t('settings.tailscale.manualIntro')}
      </p>
      <ul className="list-decimal text-left font-mono text-sm text-neutral-300">
        <li>
          {t('settings.tailscale.download')}
          <a
            className="px-1"
            href="https://pkgs.tailscale.com/stable/tailscale_latest_riscv64.tgz"
            target="_blank"
          >
            {t('settings.tailscale.package')}
          </a>
          {t('settings.tailscale.unzip')}
        </li>
        <li>{t('settings.tailscale.copyBinaries', { dir: addonDir })}</li>
        <li>
          {t('settings.tailscale.linksFile')}
          <pre className="my-1 whitespace-pre-wrap text-neutral-400">{links}</pre>
        </li>
        <li>{t('settings.tailscale.rebootRefresh')}</li>
      </ul>
    </Card>
  );
};
