import { createSignal } from 'solid-js'
import { createTenantInvite, revokeTenantInvite } from '@/services/iam'
import { tokenStorage } from '@/services/tokenStorage'
import type { TenantInviteFormValues } from '../forms'
import { roleOptions } from '../presentation'
import type { WorkspaceAccessViewModel } from '../team-access/createWorkspaceAccessViewModel'

export function createInvitesViewModel(access: WorkspaceAccessViewModel) {
    const [email, setEmail] = createSignal('')
    const [roleName, setRoleName] = createSignal(roleOptions[1].value)
    const [saving, setSaving] = createSignal(false)
    const [mutationError, setMutationError] = createSignal('')
    const [message, setMessage] = createSignal('')
    const [latestAcceptURL, setLatestAcceptURL] = createSignal('')
    const error = () => mutationError() || access.error()

    const reload = async (tenantID = access.selectedTenantID().trim()) => {
        if (!tenantID) return
        access.setSelectedTenantID(tenantID)
        await access.invites.reload()
    }

    const create = async (values: TenantInviteFormValues) => {
        const tenantID = values.tenantId.trim()
        const inviteEmail = values.email.trim()
        if (!tenantID || !inviteEmail) {
            setMutationError('Workspace id and invite email are required.')
            return
        }
        if (!access.canManage()) {
            setMutationError('You do not have permission to manage workspace invites.')
            return
        }

        setSaving(true)
        setMutationError('')
        setMessage('')
        setLatestAcceptURL('')
        access.setSelectedTenantID(tenantID)
        setEmail(inviteEmail)
        setRoleName(values.roleName)
        try {
            const result = await createTenantInvite({
                tenantId: tenantID,
                email: inviteEmail,
                roleName: values.roleName,
            })
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            setEmail('')
            setLatestAcceptURL(result.data.acceptUrl)
            setMessage(`Created a workspace invite for ${inviteEmail} in ${tenantID}.`)
            await reload(tenantID)
        } finally {
            setSaving(false)
        }
    }

    const revoke = async (inviteID: string, tenantID: string, inviteEmail: string) => {
        if (!access.canManage()) {
            setMutationError('You do not have permission to revoke workspace invites.')
            return
        }
        setMutationError('')
        setMessage('')
        const result = await revokeTenantInvite(inviteID)
        if (!result.success) {
            setMutationError(result.message)
            return
        }
        setMessage(`Revoked workspace invite for ${inviteEmail}.`)
        await reload(tenantID)
    }

    const currentUserEmail = () => tokenStorage.getUser()?.email ?? ''

    return {
        email,
        setEmail,
        roleName,
        setRoleName,
        saving,
        error,
        message,
        latestAcceptURL,
        currentUserEmail,
        reload,
        create,
        revoke,
        access,
    }
}

export type InvitesViewModel = ReturnType<typeof createInvitesViewModel>
