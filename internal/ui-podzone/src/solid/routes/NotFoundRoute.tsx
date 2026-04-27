import { useNavigate, useRouterState } from '@tanstack/solid-router';
import { NotFoundPage } from '../pages/NotFoundPage';

export default function NotFoundRoute() {
  const navigate = useNavigate();
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <NotFoundPage navigate={(to) => void navigate({ to })} path={pathname()} />
  );
}
