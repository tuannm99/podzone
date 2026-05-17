import { useNavigate } from '@tanstack/solid-router';
import { onMount, createSignal, Show } from 'solid-js';
import { Card, Button } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { ErrorAlert } from '../../components/common/Feedback';
import { tokenStorage } from '../../../services/tokenStorage';
import { tenantStorage } from '../../../services/tenantStorage';

type BootstrapBundle = {
  tenantId: string;
  tenantSlug?: string;
  username: string;
  email: string;
  password?: string;
  userId?: number;
  sessionId?: string;
  accessToken: string;
  refreshToken: string;
};

const DEV_BUNDLE_URL = '/dev-auth-bootstrap.json';

function toStoredUser(bundle: BootstrapBundle) {
  return {
    id: bundle.userId != null ? String(bundle.userId) : undefined,
    username: bundle.username,
    email: bundle.email,
  };
}

export default function DevAuthBootstrapPage() {
  const navigate = useNavigate();
  const [status, setStatus] = createSignal<'loading' | 'ready' | 'error'>(
    'loading'
  );
  const [message, setMessage] = createSignal(
    'Loading local dev auth bundle...'
  );
  const [bundle, setBundle] = createSignal<BootstrapBundle | null>(null);

  const importBundle = async () => {
    setStatus('loading');
    setMessage('Loading local dev auth bundle...');

    try {
      const response = await fetch(DEV_BUNDLE_URL, { cache: 'no-store' });
      if (!response.ok) {
        throw new Error(
          response.status === 404
            ? 'No dev auth bundle found in the UI public assets.'
            : `Failed to load dev auth bundle (${response.status}).`
        );
      }

      const nextBundle = (await response.json()) as BootstrapBundle;
      if (
        !nextBundle.accessToken ||
        !nextBundle.refreshToken ||
        !nextBundle.tenantId
      ) {
        throw new Error('Dev auth bundle is missing required fields.');
      }

      tokenStorage.clearAll();
      tenantStorage.clearTenantID();
      tokenStorage.setToken(nextBundle.accessToken);
      tokenStorage.setRefreshToken(nextBundle.refreshToken);
      tokenStorage.setUser(toStoredUser(nextBundle));
      tenantStorage.setTenantID(nextBundle.tenantId);
      setBundle(nextBundle);
      setStatus('ready');
      setMessage('Dev credentials imported. Redirecting to tenant workspace...');

      window.setTimeout(() => {
        void navigate({ to: '/t/$tenantId', params: { tenantId: nextBundle.tenantId }, replace: true });
      }, 250);
    } catch (error) {
      setStatus('error');
      setBundle(null);
      setMessage(
        error instanceof Error
          ? error.message
          : 'Unable to import the dev auth bundle.'
      );
    }
  };

  onMount(() => {
    void importBundle();
  });

  return (
    <div class="mx-auto flex min-h-[calc(100vh-2rem)] max-w-4xl items-center px-4 py-8 sm:px-5 lg:px-6">
      <Card class="w-full space-y-6 bg-[linear-gradient(135deg,_#e0f2fe,_#ffffff_42%,_#f8fafc)]">
        <SectionLead
          eyebrow="DEV BOOTSTRAP"
          title="Import local POD dev credentials"
          copy="This page reads the generated auth bundle from the UI public assets, writes it into local storage, and opens the seeded tenant workspace."
        />

        <div class="grid gap-4 rounded-2xl border border-sky-100 bg-white/85 p-5 shadow-sm">
          <p class="text-sm text-gray-600">{message()}</p>

          <Show when={status() === 'ready' && bundle()}>
            {(resolved) => (
              <div class="grid gap-2 rounded-2xl bg-slate-50 p-4 text-sm text-slate-700">
                <p>
                  <span class="font-semibold text-slate-900">Tenant:</span>{' '}
                  {resolved().tenantId}
                </p>
                <p>
                  <span class="font-semibold text-slate-900">User:</span>{' '}
                  {resolved().username}
                </p>
                <p>
                  <span class="font-semibold text-slate-900">Email:</span>{' '}
                  {resolved().email}
                </p>
              </div>
            )}
          </Show>

          <Show when={status() === 'error'}>
            <div class="space-y-4">
              <ErrorAlert>{message()}</ErrorAlert>
              <div class="rounded-2xl bg-slate-50 p-4 text-sm text-slate-700">
                <p class="font-semibold text-slate-900">Expected flow</p>
                <p class="mt-2">
                  Run `make dev-pod-sample ...` first, then sync the generated
                  bundle into the UI assets with `make dev-ui-auth-sync`.
                </p>
              </div>
            </div>
          </Show>

          <div class="flex flex-col gap-3 sm:flex-row">
            <Button onClick={() => void importBundle()} loading={status() === 'loading'}>
              Retry import
            </Button>
            <Button href="/auth/login" color="alternative">
              Back to sign in
            </Button>
            <a
              href={DEV_BUNDLE_URL}
              class="inline-flex items-center justify-center rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
            >
              Inspect bundle JSON
            </a>
          </div>
        </div>
      </Card>
    </div>
  );
}
