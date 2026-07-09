import { AdminProvisioningView } from './admin-provisioning/AdminProvisioningView'
import { createConnectionsViewModel } from './admin-provisioning/connections/createConnectionsViewModel'
import { AdminProvisioningContext, type AdminProvisioningViewModel } from './admin-provisioning/context'
import { createPipelineViewModel } from './admin-provisioning/pipeline/createPipelineViewModel'
import { createProvisioningShellViewModel } from './admin-provisioning/createProvisioningShellViewModel'
import { createResourcesViewModel } from './admin-provisioning/resources/createResourcesViewModel'

function createAdminProvisioningViewModel(): AdminProvisioningViewModel {
    const shell = createProvisioningShellViewModel()
    return {
        shell,
        pipeline: createPipelineViewModel(shell.selectedTenantId, shell.workspaceReady),
        resources: createResourcesViewModel(),
        connections: createConnectionsViewModel(shell.selectedTenantId, shell.workspaceReady),
    }
}

export default function AdminProvisioningPage() {
    const viewModel = createAdminProvisioningViewModel()
    return (
        <AdminProvisioningContext.Provider value={viewModel}>
            <AdminProvisioningView />
        </AdminProvisioningContext.Provider>
    )
}
