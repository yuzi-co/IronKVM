// What GET /api/vm/ssh and /api/vm/ssh/keys return.

export type SshKey = {
  type: string;
  comment: string;
  fingerprint: string;
};

export type RootPassword = 'default' | 'empty' | 'set' | 'unknown';

export type SshState = {
  enabled: boolean;
  running: boolean;
  port: number;
  keysOnly: boolean;
  // sshd's own answer: 'yes', 'no', or '' when it could not be asked.
  passwordAuth: string;
  keyCount: number;
  hostKeys: SshKey[] | null;
  rootPassword: RootPassword;
};
