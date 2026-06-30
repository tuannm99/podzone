import { listAuditLogs, type AuditLogInfo } from '@/services/auth'
import { createPaginatedResource } from '../shared/createPaginatedResource'

export function createAuditViewModel() {
  return createPaginatedResource<AuditLogInfo>(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'created_at',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listAuditLogs(query)
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
}

export type AuditViewModel = ReturnType<typeof createAuditViewModel>
