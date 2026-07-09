import { Link } from '@tanstack/solid-router'
import { Button, Card, InputField } from '@podzone/shared/ui/components/common/Primitives'
import { ErrorAlert } from '@podzone/shared/ui/components/common/Feedback'
import { SectionLead } from '@podzone/shared/ui/components/common/SectionLead'
import { createRegisterViewModel } from './createRegisterViewModel'

export default function RegisterPage() {
    const vm = createRegisterViewModel()

    return (
        <div class="mx-auto flex min-h-[calc(100vh-3rem)] max-w-6xl items-center px-4 py-10 sm:px-6 lg:px-8">
            <div class="grid w-full gap-6 lg:grid-cols-[0.9fr_1.1fr]">
                <Card class="space-y-6 bg-gray-950 text-white [&_h1]:text-white [&_p]:text-gray-300">
                    <SectionLead
                        eyebrow="Store Setup"
                        title="Create your seller account."
                        copy="Start with one account, then create stores, invite teammates, and run day-to-day POD operations from the same backoffice."
                    />
                    <div class="rounded-lg border border-white/10 bg-white/10 p-5 text-sm text-gray-200 shadow-sm">
                        New accounts go straight into the same secure sign-in flow used by the rest of the backoffice.
                    </div>
                </Card>

                <Card class="space-y-5">
                    <div class="space-y-1">
                        <h1 class="text-2xl font-semibold tracking-tight text-gray-900">Create account</h1>
                        <p class="text-sm text-gray-500">Set up your account to create stores and manage your team.</p>
                    </div>

                    <form class="space-y-4" onSubmit={vm.submit}>
                        <InputField
                            label="Username"
                            value={vm.form.values.username}
                            placeholder="store_owner"
                            onInput={(event) => vm.form.setValue('username', event.currentTarget.value)}
                        />
                        <InputField
                            label="Email"
                            type="email"
                            value={vm.form.values.email}
                            placeholder="owner@store.com"
                            onInput={(event) => vm.form.setValue('email', event.currentTarget.value)}
                        />
                        <InputField
                            label="Password"
                            type="password"
                            value={vm.form.values.password}
                            placeholder="Create a password"
                            onInput={(event) => vm.form.setValue('password', event.currentTarget.value)}
                        />
                        <InputField
                            label="Confirm password"
                            type="password"
                            value={vm.form.values.confirmPassword}
                            placeholder="Repeat the password"
                            error={vm.form.hasError('confirmPassword')}
                            errorText={vm.form.error('confirmPassword')}
                            onInput={(event) => vm.form.setValue('confirmPassword', event.currentTarget.value)}
                        />

                        {vm.error() ? <ErrorAlert>{vm.error()}</ErrorAlert> : null}

                        <Button
                            type="submit"
                            loading={vm.form.isSubmitting()}
                            disabled={
                                !vm.form.values.username.trim() ||
                                !vm.form.values.email.trim() ||
                                !vm.form.values.password ||
                                !vm.form.values.confirmPassword
                            }
                        >
                            Create account
                        </Button>
                    </form>

                    <p class="text-sm text-gray-500">
                        Already have an account?{' '}
                        <Link to="/auth/login" class="font-semibold text-gray-950">
                            Sign in
                        </Link>
                    </p>
                </Card>
            </div>
        </div>
    )
}
