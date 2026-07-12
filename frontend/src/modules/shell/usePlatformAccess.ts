import { createResource } from 'solid-js'
import { listOrganizations } from '@podzone/shared/services/iam'
import { tokenStorage } from '@podzone/shared/services/tokenStorage'

// Single source for the two signals that decide platform vs org nav
// visibility. listOrganizations already computes both server-side
// (internal/iam/controller/grpchandler/organization_methods.go): platform
// admins get every organization, everyone else gets only orgs they belong
// to — so canManagePlatform and "member of at least one org" both come out
// of one call, no new endpoint needed.
export function usePlatformAccess() {
    const [access] = createResource(
        () => (tokenStorage.getToken() ? true : undefined),
        async () => {
            const result = await listOrganizations({ page: 1, pageSize: 1 })
            if (!result.success) {
                return { canManagePlatform: false, hasAnyOrganization: false }
            }
            return {
                canManagePlatform: result.data.canManagePlatform,
                hasAnyOrganization: result.data.pageInfo.total > 0,
            }
        }
    )

    return {
        canManagePlatform: () => access()?.canManagePlatform ?? false,
        hasAnyOrganization: () => access()?.hasAnyOrganization ?? false,
        loading: () => access.loading,
    }
}
