import { createSignal } from 'solid-js'
import { useAuthContext } from '@podzone/shared/auth'
import { createAdminIamActions } from './createAdminIamActions'
import { createAdminIamContexts } from './createAdminIamContexts'
import { createAdminIamLoaders } from './createAdminIamLoaders'
import { createAdminIamResources } from './createAdminIamResources'
import { createAdminIamState } from './createAdminIamState'
import type { IamSectionID } from './presentation'

export function createAdminIamViewModel() {
    const auth = useAuthContext()
    const userID = auth.getUserId() || 0

    const [activeSection, setActiveSectionSignal] = createSignal<IamSectionID>('iam-orgs')
    const setSection = (section: IamSectionID) => {
        setActiveSectionSignal(section)
    }

    const state = createAdminIamState(userID, activeSection)
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
        activeSection,
        setSection,
        ...createAdminIamContexts(state, loaders, actions),
    }
}

export type AdminIamViewModel = ReturnType<typeof createAdminIamViewModel>
