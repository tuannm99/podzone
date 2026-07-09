import { useAuthContext } from '@/modules/shell/auth-context'
import { createAdminIamActions } from './createAdminIamActions'
import { createAdminIamContexts } from './createAdminIamContexts'
import { createAdminIamLoaders } from './createAdminIamLoaders'
import { createAdminIamResources } from './createAdminIamResources'
import { createAdminIamState } from './createAdminIamState'

export function createAdminIamViewModel() {
    const auth = useAuthContext()
    const userID = auth.getUserId() || 0
    const state = createAdminIamState(userID)
    const loaders = createAdminIamLoaders(state, userID)
    const actions = createAdminIamActions(state, loaders)
    createAdminIamResources(state, loaders)

    return {
        feedback: {
            error: state.pageError,
            message: state.pageMessage,
            loading: state.loading,
            allowed: state.allowed,
            canManagePlatform: state.canManagePlatform,
        },
        ...createAdminIamContexts(state, loaders, actions),
    }
}

export type AdminIamViewModel = ReturnType<typeof createAdminIamViewModel>
