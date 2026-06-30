import { AdminIamView } from './admin-iam/AdminIamView'
import { createAdminIamViewModel } from './admin-iam/createAdminIamViewModel'

export default function AdminIamPage() {
  const viewModel = createAdminIamViewModel()
  return <AdminIamView model={viewModel} />
}
