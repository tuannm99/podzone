import { createContext, useContext } from 'solid-js'
import type { ConnectionsViewModel } from './connections/createConnectionsViewModel'
import type { PipelineViewModel } from './pipeline/createPipelineViewModel'
import type { ProvisioningShellViewModel } from './createProvisioningShellViewModel'
import type { ResourcesViewModel } from './resources/createResourcesViewModel'

export type AdminProvisioningViewModel = {
    shell: ProvisioningShellViewModel
    pipeline: PipelineViewModel
    resources: ResourcesViewModel
    connections: ConnectionsViewModel
}

export const AdminProvisioningContext = createContext<AdminProvisioningViewModel>()

export function useAdminProvisioning() {
    const value = useContext(AdminProvisioningContext)
    if (!value) throw new Error('AdminProvisioningContext is missing')
    return value
}
