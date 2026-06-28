import { createResource, type Accessor, type Setter } from 'solid-js'
import {
  listAuditLogs,
  listSessions,
  revokeSession,
  type AuditLogInfo,
  type SessionInfo,
} from '@/services/auth'

export function createSessionAdmin(
  sessionID: Accessor<string>,
  setPageError: Setter<string>,
  setMessage: Setter<string>
) {
  const [sessionsResource, { refetch: refetchSessions }] = createResource(
    async (): Promise<SessionInfo[]> => {
      const result = await listSessions()
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
  const [auditLogsResource, { refetch: refetchAuditLogs }] = createResource(
    async (): Promise<AuditLogInfo[]> => {
      const result = await listAuditLogs(25)
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
  const sessions = () => sessionsResource() || []
  const auditLogs = () => auditLogsResource() || []
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
    loadingSessions,
    loadingAuditLogs,
    sessionReadError,
    auditReadError,
    currentSessionCount,
    otherSessionCount,
    loadSessions,
    loadAuditLogs,
    handleRevokeSession,
  }
}
