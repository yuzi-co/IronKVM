import { TriangleAlertIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { MediaWarning } from './warnings.ts';

type MediaWarningsProps = {
  warnings: MediaWarning[];
  values: Record<string, string>;
};

// MediaWarnings draws a drive's or an image's warnings as lines under it, as
// PiKVM's Drive menu does.
export const MediaWarnings = ({ warnings, values }: MediaWarningsProps) => {
  const { t } = useTranslation();

  if (warnings.length === 0) return null;

  return (
    <div className="flex flex-col space-y-0.5 pl-6">
      {warnings.map((warning) => (
        <div key={warning} className="flex items-start space-x-1 text-xs text-amber-400">
          <TriangleAlertIcon size={12} className="mt-[2px] shrink-0" />
          <span>{t(`image.warning.${warning}`, values)}</span>
        </div>
      ))}
    </div>
  );
};
