import type { ReactNode } from 'react';

type SectionProps = {
  title: string;
  children: ReactNode;
};

export const Section = ({ title, children }: SectionProps) => (
  <section className="flex flex-col space-y-6">
    <div className="text-sm text-neutral-400">{title}</div>
    {children}
  </section>
);
