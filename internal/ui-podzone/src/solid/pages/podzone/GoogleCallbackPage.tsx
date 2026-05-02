import { useNavigate, useSearch } from '@tanstack/solid-router';
import { createResource, createSignal } from 'solid-js';
import { exchangeGoogleLogin } from '../../../services/auth';
import { ErrorAlert, LoadingBlock } from '../../components/common/Feedback';
import { Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';

export default function GoogleCallbackPage() {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as () => Record<string, unknown>;
  const [error, setError] = createSignal('');

  createResource(async () => {
    const rawExchangeCode = search().exchange_code;
    const exchangeCode =
      typeof rawExchangeCode === 'string' ? rawExchangeCode : '';
    if (!exchangeCode) {
      setError('Missing Google sign-in exchange code.');
      return;
    }
    const result = await exchangeGoogleLogin(exchangeCode);
    if (!result.success) {
      setError(result.data.message || 'Google sign-in failed');
      return;
    }
    void navigate({ to: '/admin', replace: true });
  });

  return (
    <div class="mx-auto flex min-h-[calc(100vh-3rem)] max-w-3xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <Card class="w-full space-y-5">
        <SectionLead
          eyebrow="Google Sign-In"
          title="Finishing secure sign-in."
          copy="The browser is exchanging a one-time code for a secure session so you can enter the backoffice without exposing tokens in the callback URL."
        />
        {error() ? <ErrorAlert>{error()}</ErrorAlert> : <LoadingBlock label="Finishing Google sign-in..." />}
      </Card>
    </div>
  );
}
