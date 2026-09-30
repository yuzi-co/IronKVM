import icon from '@/assets/images/tailscale.svg';

export const Tailscale = ({ size = 18 }: { size?: number }) => {
  return <img src={icon} width={size} height={size} alt="tailscale" />;
};
