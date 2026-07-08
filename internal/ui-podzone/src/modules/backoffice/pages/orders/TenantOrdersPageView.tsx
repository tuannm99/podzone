import { TenantOrdersView } from './TenantOrdersView'
import { createTenantOrdersViewModel } from './createTenantOrdersViewModel'

export function TenantOrdersPageView() {
    const viewModel = createTenantOrdersViewModel()
    return <TenantOrdersView {...viewModel} />
}
