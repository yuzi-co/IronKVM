import { useEffect, useState } from 'react';
import { FileTextIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { getLicenseFileUrl, listLicenses } from '@/api/licenses.ts';
import { groupLicenseFiles, LicenseGroups } from '@/lib/licenses.ts';

import { SectionHeader } from '../components/section.tsx';

// Licenses links the files of the image's license manifest, each opening as
// plain text in a new tab. An image without a manifest (a development build)
// or an older server answers 404, and the section says so in grey instead.
export const Licenses = () => {
  const { t } = useTranslation();

  const [groups, setGroups] = useState<LicenseGroups>();
  const [unavailable, setUnavailable] = useState(false);

  useEffect(() => {
    listLicenses()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setUnavailable(true);
          return;
        }
        setGroups(groupLicenseFiles(rsp.data?.files));
      })
      .catch(() => setUnavailable(true));
  }, []);

  const empty = groups && groups.summary.length === 0 && groups.texts.length === 0;

  return (
    <>
      <SectionHeader
        title={t('settings.about.licenses')}
        description={t('settings.about.licensesDescription')}
      />

      {unavailable || empty ? (
        <div className="mt-4 text-xs text-neutral-500">
          {t('settings.about.licensesUnavailable')}
        </div>
      ) : (
        groups && (
          <div className="mt-4 flex flex-col space-y-3">
            <div className="flex flex-wrap gap-2">
              {groups.summary.map((file) => (
                <a
                  key={file.path}
                  className="flex items-center space-x-1 rounded px-2 py-1 font-mono text-xs text-neutral-300 outline outline-1 outline-neutral-800 hover:bg-neutral-800 hover:text-white"
                  href={getLicenseFileUrl(file.path)}
                  target="_blank"
                  rel="noreferrer"
                >
                  <FileTextIcon size={14} />
                  <span>{file.path}</span>
                </a>
              ))}
            </div>

            {groups.texts.length > 0 && (
              <details className="text-xs text-neutral-400">
                <summary className="cursor-pointer select-none hover:text-neutral-200">
                  {t('settings.about.licenseTexts', { count: groups.texts.length })}
                </summary>
                <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1">
                  {groups.texts.map((text) => (
                    <a
                      key={text.path}
                      className="font-mono text-neutral-300 hover:text-blue-500"
                      href={getLicenseFileUrl(text.path)}
                      target="_blank"
                      rel="noreferrer"
                    >
                      {text.name}
                    </a>
                  ))}
                </div>
              </details>
            )}
          </div>
        )
      )}
    </>
  );
};
