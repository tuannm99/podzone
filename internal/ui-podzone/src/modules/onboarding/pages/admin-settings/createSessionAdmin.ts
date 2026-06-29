import { createResource, type Accessor, type Setter } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  listAuditLogs,
  listSessions,
  revokeSession,
  type AuditLogInfo,
  type SessionInfo,
} from '@/services/auth'
import {
  emptyPageInfo,
  type CollectionPage,
  type CollectionQuery,
} from '@/services/collection'

export function createSessionAdmin(
  sessionID: Accessor<string>,
  setPageError: Setter<string>,
  setMessage: Setter<string>
) {
  const [sessionQuery, setSessionQuery] = createStore<CollectionQuery>({
    page: 1,
    pageSize: 5,
    sortBy: 'created_at',
    sortDirection: 'SORT_DIRECTION_DESC',
  })
  const [auditQuery, setAuditQuery] = createStore<CollectionQuery>({
    page: 1,
    pageSize: 10,
    sortBy: 'created_at',
    sortDirection: 'SORT_DIRECTION_DESC',
  })
  const [sessionsResource, { refetch: refetchSessions }] = createResource(
    () => ({ ...sessionQuery }),
    async (query): Promise<CollectionPage<SessionInfo>> => {
      const result = await listSessions(query)
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
  const [auditLogsResource, { refetch: refetchAuditLogs }] = createResource(
    () => ({ ...auditQuery }),
    async (query): Promise<CollectionPage<AuditLogInfo>> => {
      const result = await listAuditLogs(query)
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
  const sessions = () => sessionsResource()?.items || []
  const auditLogs = () => auditLogsResource()?.items || []
  const sessionPageInfo = () =>
    sessionsResource()?.pageInfo || emptyPageInfo(sessionQuery)
  const auditPageInfo = () =>
    auditLogsResource()?.pageInfo || emptyPageInfo(auditQuery)
  const loadingSessions = () => sessionsResource.loading
  const loadingAuditLogs = () => auditLogsResource.loading
  const sessionReadError = () =>
    sessionsResource.error instanceof Error
      ? sessionsResource.error.message
      : ''
  const auditReadError = () =>
    auditLogsResource.error instanceof Error
      ? auditLogsResource.error.message
      : ''
  const currentSessionCount = () =>
    sessions().filter((session) => session.id === sessionID()).length
  const otherSessionCount = () =>
    sessions().filter((session) => session.id !== sessionID()).length

  const loadSessions = async () => void (await refetchSessions())
  const loadAuditLogs = async () => void (await refetchAuditLogs())
  const updateSessionQuery = (patch: Partial<CollectionQuery>) => {
    setSessionQuery({ ...patch, page: patch.page ?? 1 })
  }
  const updateAuditQuery = (patch: Partial<CollectionQuery>) => {
    setAuditQuery({ ...patch, page: patch.page ?? 1 })
  }

  const handleRevokeSession = async (sessionId: string) => {
    setPageError('')
    setMessage('')
    const result = await revokeSession(sessionId)
    if (!result.success) {
      setPageError(result.data.message || 'Failed to revoke session')
      return
    }
    setMessage(`Revoked session ${sessionId}.`)
    await loadSessions()
  }

  return {
    sessions,
    auditLogs,
    sessionQuery,
    auditQuery,
    sessionPageInfo,
    auditPageInfo,
    loadingSessions,
    loadingAuditLogs,
    sessionReadError,
    auditReadError,
    currentSessionCount,
    otherSessionCount,
    loadSessions,
    loadAuditLogs,
    updateSessionQuery,
    updateAuditQuery,
    handleRevokeSession,
  }
}
