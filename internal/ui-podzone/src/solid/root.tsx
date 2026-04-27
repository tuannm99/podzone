import { Outlet, useRouterState } from '@tanstack/solid-router';
import { Show, createEffect } from 'solid-js';
import { AppShell } from './components/common/AppShell';
import { ScrollToTopButton } from './components/common/ScrollToTop';
import { PodzoneNavbar } from './layout/PodzoneNavbar';

export default function Root() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const isAuthRoute = () => pathname().startsWith('/auth/');

  createEffect(() => {
    pathname();
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto',
    });
  });

  return (
    <AppShell class="bg-[radial-gradient(circle_at_top,_rgba(191,219,254,0.35),_transparent_42%),linear-gradient(180deg,_#f8fafc,_#eef2ff_42%,_#f8fafc)]">
      <Show when={!isAuthRoute()}>
        <PodzoneNavbar currentPath={pathname()} />
      </Show>

      <main class="mt-4 pb-10">
        <Outlet />
      </main>

      <Show when={!isAuthRoute()}>
        <ScrollToTopButton />
      </Show>
    </AppShell>
  );
}
