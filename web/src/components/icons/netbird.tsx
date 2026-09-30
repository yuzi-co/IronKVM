import icon from '@/assets/images/netbird.svg';

export const Netbird = ({ size = 18 }: { size?: number }) => {
  return <img src={icon} width={size} height={size} className="object-contain" alt="netbird" />;
};
