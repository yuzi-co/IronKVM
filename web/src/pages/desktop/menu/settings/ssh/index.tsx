import { Divider } from 'antd';

import { Ssh as SshSwitch } from '../device/ssh.tsx';

// SSH has one switch. It used to sit in Device; it lives with the other ways
// in to the device now, and the component stays where it was.
export const Ssh = () => (
  <>
    <div className="text-base">SSH</div>
    <Divider className="opacity-50" />

    <SshSwitch />
  </>
);
