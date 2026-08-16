import { GithubOutlined } from '@ant-design/icons';
import { BookOpenIcon, CpuIcon, MessageCircleQuestionIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export const Community = () => {
  const { t } = useTranslation();

  // Sipeed's own X and Discord are deliberately absent. They are where the
  // hardware vendor supports its own firmware, and sending a report about this
  // one there costs somebody else time for a build they did not make. The
  // hardware links stay, because a hardware question still belongs to them.
  const communities = [
    {
      name: 'Document',
      icon: <BookOpenIcon size={24} />,
      url: 'https://github.com/yuzi-co/IronKVM#readme'
    },
    {
      name: 'GitHub',
      icon: <GithubOutlined style={{ fontSize: '20px' }} width={24} height={24} />,
      url: 'https://github.com/yuzi-co/IronKVM'
    },
    {
      name: 'Hardware',
      icon: <CpuIcon size={24} />,
      url: 'https://wiki.sipeed.com/nanokvm'
    },
    {
      name: 'Hardware FAQ',
      icon: <MessageCircleQuestionIcon size={24} />,
      url: 'https://wiki.sipeed.com/hardware/en/kvm/NanoKVM/faq.html'
    }
  ];

  return (
    <>
      <div className="text-neutral-400">{t('settings.about.community')}</div>

      <div className="mt-5 flex flex-wrap gap-3">
        {communities.map((community) => (
          <a
            key={community.name}
            className="flex h-[64px] w-[80px] flex-col items-center justify-center space-y-2 rounded-lg text-neutral-300 outline outline-1 outline-neutral-800 hover:bg-neutral-800 hover:text-white focus:bg-neutral-800 md:h-[72px] md:w-[100px]"
            href={community.url}
            target="_blank"
          >
            {community.icon}
            <span className="text-xs">{community.name}</span>
          </a>
        ))}
      </div>

      {/*
        Stated here rather than only in the README, because this panel is where
        somebody looks when they want to know what they are running. The name
        deliberately carries no "Nano" and no "Sipeed", and this is the second
        half of keeping that honest.
      */}
      <div className="mt-4 text-xs text-neutral-500">
        IronKVM: hardened community firmware for the Sipeed NanoKVM. Not affiliated with Sipeed.
      </div>
    </>
  );
};
