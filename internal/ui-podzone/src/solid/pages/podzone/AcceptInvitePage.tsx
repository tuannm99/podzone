import { useNavigate, useSearch } from '@tanstack/solid-router';
import { Show, createSignal } from 'solid-js';
import { switchActiveTenant } from '../../../services/auth';
import { acceptTenantInvite } from '../../../services/iam';
import { tokenStorage } from '../../../services/tokenStorage';
import { ErrorAlert, InfoAlert } from '../../components/common/Feedback';
import { Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';

export default function AcceptInvitePage() {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as () => Record<string, unknown>;
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal('');
  const [message, setMessage] = createSignal('');

  const inviteToken = () => {
    const rawToken = search().token;
    return typeof rawToken === 'string' ? rawToken.trim() : '';
  };

  const isAuthenticated = () => Boolean(tokenStorage.getToken());

  const acceptInvite = async () => {
    const token = inviteToken();
    if (!token) {
      setError('Missing store invite token.');
      return;
    }
    if (!isAuthenticated()) {
      setError(
        'Sign in with the invited account before joining this store.'
      );
      return;
    }

    setLoading(true);
    setError('');
    setMessage('');
    try {
      const accepted = await acceptTenantInvite(token);
      if (!accepted.success) {
        setError(accepted.message);
        return;
      }

      const tenantId = accepted.data.tenantId;
      const switched = await switchActiveTenant(tenantId);
      if (!switched.success) {
        setMessage(
          `Store invite accepted for ${tenantId}. Open that store manually from the backoffice if switching does not happen automatically.`
        );
        void navigate({ to: '/admin', replace: true });
        return;
      }

      setMessage(`Store invite accepted. Opening store ${tenantId}.`);
      void navigate({
        to: '/t/$tenantId',
        params: { tenantId },
        replace: true,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="mx-auto flex min-h-[calc(100vh-3rem)] max-w-3xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <Card class="w-full space-y-5">
        <SectionLead
          eyebrow="Store Invite"
          title="Join your store team."
          copy="This invite links your signed-in account to the right store role, then opens that store so you can start working right away."
        />

        <Show when={error()}>
          <ErrorAlert>{error()}</ErrorAlert>
        </Show>

        <Show when={message()}>
          <InfoAlert>{message()}</InfoAlert>
        </Show>

        <Show when={!inviteToken()}>
          <ErrorAlert>Store invite token is missing from this URL.</ErrorAlert>
        </Show>

        <Show when={!isAuthenticated()}>
          <InfoAlert>
            Sign in with the invited account, then reopen this link to finish
            joining the store.
          </InfoAlert>
        </Show>

        <div class="flex flex-wrap gap-3">
          <Button
            type="button"
            loading={loading()}
            disabled={!inviteToken() || !isAuthenticated()}
            onClick={() => {
              void acceptInvite();
            }}
          >
            Join store
          </Button>
          <Button color="alternative" href="/auth/login">
            Go to sign in
          </Button>
          <Button color="light" href="/admin">
            Back to backoffice
          </Button>
        </div>
      </Card>
    </div>
  );
}
