import { useNavigate, useSearch } from '@tanstack/solid-router'
import { Show, createSignal } from 'solid-js'
import { switchActiveTenant } from '@/services/auth'
import { acceptTenantInvite } from '@/services/iam'
import { useAuthContext } from '@/solid/context/auth-context'
import { ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'

export default function AcceptInvitePage() {
    const auth = useAuthContext()
    const navigate = useNavigate()
    const search = useSearch({ strict: false }) as () => Record<string, unknown>
    const [loading, setLoading] = createSignal(false)
    const [error, setError] = createSignal('')
    const [message, setMessage] = createSignal('')

    const inviteToken = () => {
        const rawToken = search().token
        return typeof rawToken === 'string' ? rawToken.trim() : ''
    }

    const isAuthenticated = () => auth.isAuthenticated()

    const acceptInvite = async () => {
        const token = inviteToken()
        if (!token) {
            setError('Missing workspace invite token.')
            return
        }
        if (!isAuthenticated()) {
            setError('Sign in with the invited account before joining this store.')
            return
        }

        setLoading(true)
        setError('')
        setMessage('')
        try {
            const accepted = await acceptTenantInvite(token)
            if (!accepted.success) {
                setError(accepted.message)
                return
            }

            const tenantId = accepted.data.tenantId
            const switched = await switchActiveTenant(tenantId)
            if (!switched.success) {
                setMessage(
                    `Workspace invite accepted for ${tenantId}. Open /admin and choose the store manually if switching does not happen automatically.`
                )
                void navigate({ to: '/admin', replace: true })
                return
            }

            setMessage(`Workspace invite accepted. Choose a store from /admin to continue.`)
            void navigate({ to: '/admin', replace: true })
        } finally {
            setLoading(false)
        }
    }

    return (
        <div class="mx-auto flex min-h-[calc(100vh-3rem)] max-w-3xl items-center px-4 py-10 sm:px-6 lg:px-8">
            <Card class="w-full space-y-5">
                <SectionLead
                    eyebrow="Workspace Invite"
                    title="Join your workspace team."
                    copy="This invite links your signed-in account to the right workspace role, then sends you back to /admin so you can pick the right store."
                />

                <Show when={error()}>
                    <ErrorAlert>{error()}</ErrorAlert>
                </Show>

                <Show when={message()}>
                    <InfoAlert>{message()}</InfoAlert>
                </Show>

                <Show when={!inviteToken()}>
                    <ErrorAlert>Workspace invite token is missing from this URL.</ErrorAlert>
                </Show>

                <Show when={!isAuthenticated()}>
                    <InfoAlert>
                        Sign in with the invited account, then reopen this link to finish joining the workspace.
                    </InfoAlert>
                </Show>

                <div class="flex flex-wrap gap-3">
                    <Button
                        type="button"
                        loading={loading()}
                        disabled={!inviteToken() || !isAuthenticated()}
                        onClick={() => {
                            void acceptInvite()
                        }}
                    >
                        Join workspace
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
    )
}
