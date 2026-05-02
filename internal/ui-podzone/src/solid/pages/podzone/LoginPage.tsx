import { Link, useNavigate } from '@tanstack/solid-router';
import { createSignal } from 'solid-js';
import { login, loginGG } from '../../../services/auth';
import { Card, Button, InputField } from '../../components/common/Primitives';
import { ErrorAlert } from '../../components/common/Feedback';
import { SectionLead } from '../../components/common/SectionLead';

export default function LoginPage() {
  const navigate = useNavigate();
  const [username, setUsername] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [error, setError] = createSignal('');
  const [loading, setLoading] = createSignal(false);

  const submit = async (event: SubmitEvent) => {
    event.preventDefault();
    setLoading(true);
    setError('');

    try {
      const { success, data } = await login({
        username: username().trim(),
        password: password(),
      });

      if (!success) {
        setError(data?.message || 'Sign-in failed');
        return;
      }

      void navigate({ to: '/admin', replace: true });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="mx-auto flex min-h-[calc(100vh-2rem)] max-w-5xl items-center px-4 py-8 sm:px-5 lg:px-6">
      <div class="grid w-full gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-6 bg-[linear-gradient(135deg,_#dbeafe,_#ffffff_45%,_#eff6ff)]">
          <SectionLead
            eyebrow="PODZONE"
            title="Run your POD stores from one backoffice."
            copy="Sign in once to manage stores, team access, invites, and the operational side of your POD business."
          />

          <div class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-2xl bg-white/80 p-4 shadow-sm">
              <p class="text-xs font-semibold uppercase tracking-[0.2em] text-blue-600">
                01
              </p>
              <p class="mt-2 text-sm text-gray-600">
                One sign-in unlocks your store operations workspace.
              </p>
            </div>
            <div class="rounded-2xl bg-white/80 p-4 shadow-sm">
              <p class="text-xs font-semibold uppercase tracking-[0.2em] text-blue-600">
                02
              </p>
              <p class="mt-2 text-sm text-gray-600">
                Store switching keeps each session scoped to the right shop.
              </p>
            </div>
            <div class="rounded-2xl bg-white/80 p-4 shadow-sm">
              <p class="text-xs font-semibold uppercase tracking-[0.2em] text-blue-600">
                03
              </p>
              <p class="mt-2 text-sm text-gray-600">
                Team access and invites stay tied to the correct store.
              </p>
            </div>
          </div>
        </Card>

        <Card class="space-y-5">
          <div class="space-y-1">
            <h1 class="text-2xl font-semibold tracking-tight text-gray-900">
              Sign in
            </h1>
            <p class="text-sm text-gray-500">
              Use your account to open the seller backoffice.
            </p>
          </div>

          <form class="space-y-4" onSubmit={submit}>
            <InputField
              label="Username or email"
              value={username()}
              placeholder="owner@store.com"
              onInput={(event) => setUsername(event.currentTarget.value)}
            />
            <InputField
              label="Password"
              type="password"
              value={password()}
              placeholder="Enter your password"
              onInput={(event) => setPassword(event.currentTarget.value)}
            />

            {error() ? <ErrorAlert>{error()}</ErrorAlert> : null}

            <div class="flex flex-col gap-3">
              <Button
                type="submit"
                loading={loading()}
                disabled={!username().trim() || !password()}
              >
                Continue
              </Button>
              <Button href={loginGG()} color="alternative">
                Sign in with Google
              </Button>
            </div>
          </form>

          <p class="text-sm text-gray-500">
            Need an account?{' '}
            <Link to="/auth/register" class="font-semibold text-blue-700">
              Create one
            </Link>
          </p>
        </Card>
      </div>
    </div>
  );
}
