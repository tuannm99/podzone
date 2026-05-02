import { Link, useNavigate } from '@tanstack/solid-router';
import { createSignal } from 'solid-js';
import { register } from '../../../services/auth';
import { Button, Card, InputField } from '../../components/common/Primitives';
import { ErrorAlert } from '../../components/common/Feedback';
import { SectionLead } from '../../components/common/SectionLead';

export default function RegisterPage() {
  const navigate = useNavigate();
  const [username, setUsername] = createSignal('');
  const [email, setEmail] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [confirmPassword, setConfirmPassword] = createSignal('');
  const [error, setError] = createSignal('');
  const [loading, setLoading] = createSignal(false);

  const submit = async (event: SubmitEvent) => {
    event.preventDefault();

    if (password() !== confirmPassword()) {
      setError('The passwords do not match.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const { success, data } = await register({
        username: username().trim(),
        email: email().trim(),
        password: password(),
      });

      if (!success) {
        setError(data?.message || 'Account setup failed');
        return;
      }

      void navigate({ to: '/admin', replace: true });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="mx-auto flex min-h-[calc(100vh-3rem)] max-w-6xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <div class="grid w-full gap-6 lg:grid-cols-[0.9fr_1.1fr]">
        <Card class="space-y-6 bg-[linear-gradient(145deg,_#ecfeff,_#ffffff_45%,_#eff6ff)]">
          <SectionLead
            eyebrow="Store Setup"
            title="Create your seller account."
            copy="Start with one account, then create stores, invite teammates, and run day-to-day POD operations from the same backoffice."
          />
          <div class="rounded-2xl border border-cyan-100 bg-white/80 p-5 text-sm text-gray-600 shadow-sm">
            New accounts go straight into the same secure sign-in flow used by
            the rest of the backoffice.
          </div>
        </Card>

        <Card class="space-y-5">
          <div class="space-y-1">
            <h1 class="text-2xl font-semibold tracking-tight text-gray-900">
              Create account
            </h1>
            <p class="text-sm text-gray-500">
              Set up your account to create stores and manage your team.
            </p>
          </div>

          <form class="space-y-4" onSubmit={submit}>
            <InputField
              label="Username"
              value={username()}
              placeholder="store_owner"
              onInput={(event) => setUsername(event.currentTarget.value)}
            />
            <InputField
              label="Email"
              type="email"
              value={email()}
              placeholder="owner@store.com"
              onInput={(event) => setEmail(event.currentTarget.value)}
            />
            <InputField
              label="Password"
              type="password"
              value={password()}
              placeholder="Create a password"
              onInput={(event) => setPassword(event.currentTarget.value)}
            />
            <InputField
              label="Confirm password"
              type="password"
              value={confirmPassword()}
              placeholder="Repeat the password"
              onInput={(event) => setConfirmPassword(event.currentTarget.value)}
            />

            {error() ? <ErrorAlert>{error()}</ErrorAlert> : null}

            <Button
              type="submit"
              loading={loading()}
              disabled={
                !username().trim() ||
                !email().trim() ||
                !password() ||
                !confirmPassword()
              }
            >
              Create account
            </Button>
          </form>

          <p class="text-sm text-gray-500">
            Already have an account?{' '}
            <Link to="/auth/login" class="font-semibold text-blue-700">
              Sign in
            </Link>
          </p>
        </Card>
      </div>
    </div>
  );
}
