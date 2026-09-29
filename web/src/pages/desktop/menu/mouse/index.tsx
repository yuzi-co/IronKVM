import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Divider } from 'antd';
import { useSetAtom } from 'jotai';
import { MouseIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as ls from '@/lib/localstorage';
import {
  mouseModeAtom,
  mouseStyleAtom,
  scrollDirectionAtom,
  scrollIntervalAtom
} from '@/jotai/mouse';
import { MenuItem } from '@/components/menu-item.tsx';

import { Cursor } from './cursor.tsx';
import { Direction } from './direction.tsx';
import { Jiggler } from './jiggler.tsx';
import { MouseMode } from './mouse-mode.tsx';
import { Speed } from './speed.tsx';

// HID mode and Reset HID are device-wide and can drop input, so they are
// settings rather than menu entries; their components stay in this folder.
// Original resolution changes what you see, so it sits in the Screen menu.
export const Mouse = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  // The server takes jiggler changes from admins only.
  const isAdmin = account.role === 'admin';
  const [openCount, setOpenCount] = useState(0);

  const setMouseStyle = useSetAtom(mouseStyleAtom);
  const setMouseMode = useSetAtom(mouseModeAtom);
  const setScrollDirection = useSetAtom(scrollDirectionAtom);
  const setScrollInterval = useSetAtom(scrollIntervalAtom);

  useEffect(() => {
    const mouseStyle = ls.getMouseStyle();
    if (mouseStyle) {
      setMouseStyle(mouseStyle);
    }

    const mouseMode = ls.getMouseMode();
    if (mouseMode) {
      setMouseMode(mouseMode);
    }

    const direction = ls.getMouseScrollDirection();
    if (direction) {
      setScrollDirection(direction > 0 ? 1 : -1);
    }

    const interval = ls.getMouseScrollInterval();
    if (interval) {
      setScrollInterval(interval);
    }
  }, [setMouseStyle, setMouseMode, setScrollDirection, setScrollInterval]);

  const content = (
    <div className="flex flex-col space-y-1">
      <Cursor />
      <MouseMode />
      <Direction />
      <Speed />
      {isAdmin && (
        <>
          <Divider style={{ margin: '10px 0' }} />
          <Jiggler refreshKey={openCount} />
        </>
      )}
    </div>
  );

  return (
    <MenuItem
      title={t('mouse.title')}
      icon={<MouseIcon size={18} />}
      content={content}
      onOpenChange={(open) => open && setOpenCount((count) => count + 1)}
    />
  );
};
