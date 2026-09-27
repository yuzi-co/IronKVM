import { Card } from 'antd';
import { useTranslation } from 'react-i18next';

// InstallHelp is the manual install, shown when the install from the page
// fails.
export const InstallHelp = () => {
  const { t } = useTranslation();

  return (
    <Card styles={{ body: { padding: 0 } }}>
      <p className="px-4 pt-3 text-left text-sm text-neutral-400">
        {t('settings.tailscale.retry')}
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
        <li>{t('settings.tailscale.upTailscale')}</li>
        <li>{t('settings.tailscale.upTailscaled')}</li>
        <li>{t('settings.tailscale.refresh')}</li>
      </ul>
    </Card>
  );
};
