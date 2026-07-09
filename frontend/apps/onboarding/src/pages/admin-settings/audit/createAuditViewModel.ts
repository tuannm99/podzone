import { listAuditLogs, type AuditLogInfo } from '@podzone/shared/services/auth'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'
import type { Accessor } from 'solid-js'

export function createAuditViewModel(enabled: Accessor<boolean>) {
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
        },
        { enabled }
    )
}

export type AuditViewModel = ReturnType<typeof createAuditViewModel>
