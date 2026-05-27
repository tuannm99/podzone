import { Outlet, useRouterState } from '@tanstack/solid-router';
import { Match, Switch, createEffect, createMemo } from 'solid-js';
import { AppShell, Container } from './components/common/AppShell';
import { ScrollToTopButton } from './components/common/ScrollToTop';
import { PodzoneNavbar } from './layout/PodzoneNavbar';
import { TenantWorkspaceProvider } from './workspace/context';

function parseTenantId(pathname: string) {
  const match = pathname.match(/^\/t\/([^/]+)/);
  return match?.[1] || '';
}

export default function Root() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const isAuthRoute = () => pathname().startsWith('/auth/');
  const tenantId = createMemo(() => parseTenantId(pathname()));

  createEffect(() => {
    pathname();
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto',
    });
  });

  return (
    <AppShell class="bg-gray-50">
      <Switch fallback={<Outlet />}>
        <Match when={isAuthRoute()}>
          <Outlet />
        </Match>
        <Match when={tenantId()}>
          <TenantWorkspaceProvider tenantId={tenantId()}>
            <PodzoneNavbar currentPath={pathname()} />
            <main class="pb-8 lg:pl-64">
              <Container class="mt-5" width="7xl">
                <div class="grid min-h-0 grid-cols-1 gap-0 xl:grid-cols-[minmax(0,1fr)]">
                  <Outlet />
                </div>
              </Container>
            </main>
            <ScrollToTopButton />
          </TenantWorkspaceProvider>
        </Match>
        <Match when={true}>
          <PodzoneNavbar currentPath={pathname()} />
          <main class="pb-8 lg:pl-64">
            <Container class="mt-5" width="7xl">
              <div class="grid min-h-0 grid-cols-1 gap-0 xl:grid-cols-[minmax(0,1fr)]">
                <Outlet />
              </div>
            </Container>
          </main>
          <ScrollToTopButton />
        </Match>
      </Switch>
    </AppShell>
  );
}
