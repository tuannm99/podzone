import { Link } from '@tanstack/solid-router'
import { loginGG } from '@podzone/shared/services/auth'
import { Card, Button, InputField } from '@podzone/shared/ui/components/common/Primitives'
import { ErrorAlert } from '@podzone/shared/ui/components/common/Feedback'
import { SectionLead } from '@podzone/shared/ui/components/common/SectionLead'
import { createLoginViewModel } from './createLoginViewModel'

export default function LoginPage() {
    const vm = createLoginViewModel()

    return (
        <div class="mx-auto flex min-h-[calc(100vh-2rem)] max-w-5xl items-center px-4 py-8 sm:px-5 lg:px-6">
            <div class="grid w-full gap-5 lg:grid-cols-[1.05fr_0.95fr]">
                <Card class="space-y-6 bg-gray-950 text-white [&_h1]:text-white [&_p]:text-gray-300">
                    <SectionLead
                        eyebrow="PODZONE"
                        title="Run your POD stores from one backoffice."
                        copy="Sign in once to manage stores, team access, invites, and the operational side of your POD business."
                    />

                    <div class="grid gap-3 sm:grid-cols-3">
                        <div class="rounded-lg bg-white/80 p-4 shadow-sm">
                            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500">01</p>
                            <p class="mt-2 text-sm text-gray-600">
                                One sign-in unlocks your store operations workspace.
                            </p>
                        </div>
                        <div class="rounded-lg bg-white/80 p-4 shadow-sm">
                            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500">02</p>
                            <p class="mt-2 text-sm text-gray-600">
                                Store switching keeps each session scoped to the right shop.
                            </p>
                        </div>
                        <div class="rounded-lg bg-white/80 p-4 shadow-sm">
                            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500">03</p>
                            <p class="mt-2 text-sm text-gray-600">
                                Team access and invites stay tied to the correct store.
                            </p>
                        </div>
                    </div>
                </Card>

                <Card class="space-y-5">
                    <div class="space-y-1">
                        <h1 class="text-2xl font-semibold tracking-tight text-gray-900">Sign in</h1>
                        <p class="text-sm text-gray-500">Use your account to open the seller backoffice.</p>
                    </div>

                    <form class="space-y-4" onSubmit={vm.submit}>
                        <InputField
                            label="Username or email"
                            value={vm.form.values.username}
                            placeholder="owner@store.com"
                            onInput={(event) => vm.form.setValue('username', event.currentTarget.value)}
                        />
                        <InputField
                            label="Password"
                            type="password"
                            value={vm.form.values.password}
                            placeholder="Enter your password"
                            onInput={(event) => vm.form.setValue('password', event.currentTarget.value)}
                        />

                        {vm.error() ? <ErrorAlert>{vm.error()}</ErrorAlert> : null}

                        <div class="flex flex-col gap-3">
                            <Button
                                type="submit"
                                loading={vm.form.isSubmitting()}
                                disabled={!vm.form.values.username.trim() || !vm.form.values.password}
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
                        <Link to="/auth/register" class="font-semibold text-gray-950">
                            Create one
                        </Link>
                    </p>
                    <p class="text-sm text-gray-500">
                        Local dev seed ready?{' '}
                        <Link to="/auth/dev/bootstrap" class="font-semibold text-gray-950">
                            Import seeded credentials
                        </Link>
                    </p>
                </Card>
            </div>
        </div>
    )
}
