import { Outlet } from 'react-router';

export const Root = () => {
  return (
    <div className="h-screen w-screen">
      <Outlet />
    </div>
  );
};
