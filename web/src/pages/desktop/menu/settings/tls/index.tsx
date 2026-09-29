import { Divider } from 'antd';

import { Tls as HttpsSwitch } from '../network/tls.tsx';

type TlsProps = {
  setIsLocked: (isLocked: boolean) => void;
};

// HTTPS has its own page under Access, next to the logins it protects. The
// switch restarts the server, so it locks the modal while it waits.
export const Tls = ({ setIsLocked }: TlsProps) => (
  <>
    <div className="text-base">TLS</div>
    <Divider className="opacity-50" />

    <HttpsSwitch setIsLocked={setIsLocked} />
  </>
);
