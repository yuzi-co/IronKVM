import { Tag } from 'antd';
import { useTranslation } from 'react-i18next';

type StatusTagProps = {
  running: boolean;
  // Set while the value shown may be old because the last poll failed.
  stale?: boolean;
};

// StatusTag is the badge beside a service page's title: green Running or grey
// Off. Every service page uses it, so the same state looks the same everywhere.
export const StatusTag = ({ running, stale = false }: StatusTagProps) => {
  const { t } = useTranslation();

  return (
    <StateTag
      color={running ? 'green' : 'default'}
      label={running ? t('common.running') : t('common.off')}
      stale={stale}
    />
  );
};

type StateTagProps = {
  color: 'green' | 'gold' | 'red' | 'default';
  label: string;
  stale?: boolean;
};

// StateTag is the same badge for a page with more states than running and
// off, such as the VPN page's Connected and Needs login.
export const StateTag = ({ color, label, stale = false }: StateTagProps) => (
  <Tag color={color} className={stale ? 'me-0 opacity-50' : 'me-0'}>
    {label}
  </Tag>
);

// StaleNote says a polled value stopped updating.
export const StaleNote = () => {
  const { t } = useTranslation();
  return <span className="text-xs text-amber-500">{t('common.notUpdating')}</span>;
};
