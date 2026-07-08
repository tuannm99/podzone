import { createAdminHomeViewModel } from './admin-home/createAdminHomeViewModel'
import { AdminHomeContext } from './admin-home/context'
import { AdminHomeView } from './admin-home/AdminHomeView'

export { createAdminHomeViewModel }
export type { AdminHomeViewModel } from './admin-home/createAdminHomeViewModel'

export default function AdminHomePage() {
    const viewModel = createAdminHomeViewModel()
    return (
        <AdminHomeContext.Provider value={viewModel}>
            <AdminHomeView />
        </AdminHomeContext.Provider>
    )
}
