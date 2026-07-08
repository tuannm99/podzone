import { createEffect, createResource, on, onMount } from 'solid-js'
import type { AdminIamLoaders } from './createAdminIamLoaders'
import type { AdminIamState } from './createAdminIamState'

export function createAdminIamResources(state: AdminIamState, loaders: AdminIamLoaders) {
    createEffect(() => {
        const firstPolicy = state.policies()[0]
        if (state.allowed() && !state.selectedPolicyName() && firstPolicy) {
            state.setSelectedPolicyName(firstPolicy.name)
        }
    })

    createEffect(() => {
        const firstOrganization = state.organizations()[0]
        if (state.allowed() && !state.selectedOrgId() && firstOrganization) {
            state.setSelectedOrgId(firstOrganization.id)
        }
    })

    createResource(
        () => (state.allowed() ? state.selectedPolicyName() : undefined),
        async () => loaders.loadSelectedPolicy()
    )

    createResource(
        () => (state.allowed() ? state.selectedOrgId() : undefined),
        async () => loaders.loadSelectedOrganization()
    )

    createResource(
        () => (state.allowed() ? state.selectedGroupId() : undefined),
        async () => loaders.loadSelectedGroup()
    )

    createResource(
        () => (state.allowed() ? `${state.groupScope()}|${state.groupTenantId()}` : undefined),
        async () => loaders.loadGroupsForScope()
    )

    createEffect(
        on(
            state.canManagePlatform,
            (canManagePlatform) => {
                if (!canManagePlatform) return
                state.setPolicyScope('platform')
                state.setGroupScope('platform')
            },
            { defer: true }
        )
    )

    createEffect(
        on(
            state.selectedOrgId,
            () => {
                if (state.policyScope() === 'organization') state.setSelectedPolicyName('')
                if (state.groupScope() === 'organization') state.setSelectedGroupId('')
            },
            { defer: true }
        )
    )

    onMount(() => {
        void loaders.loadBootstrap()
    })
}
