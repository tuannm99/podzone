import { createSignal, type Accessor } from 'solid-js'
import { listSessions, revokeSession, type SessionInfo } from '@/services/auth'
import { createPaginatedResource } from '../shared/createPaginatedResource'

export function createSessionsViewModel(sessionID: Accessor<string>) {
  const list = createPaginatedResource<SessionInfo>(
    {
      page: 1,
      pageSize: 5,
      sortBy: 'created_at',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listSessions(query)
      if (!result.success) throw new Error(result.data.message)
      return result.data
    }
  )
  const [mutationError, setMutationError] = createSignal('')
  const [message, setMessage] = createSignal('')
  const currentCount = () =>
    list.items().filter((session) => session.id === sessionID()).length
  const otherCount = () =>
    list.items().filter((session) => session.id !== sessionID()).length
  const error = () => mutationError() || list.error()

  const revoke = async (sessionId: string) => {
    setMutationError('')
    setMessage('')
    const result = await revokeSession(sessionId)
    if (!result.success) {
      setMutationError(result.data.message || 'Failed to revoke session')
      return
    }
    setMessage(`Revoked session ${sessionId}.`)
    await list.reload()
  }

  return {
    query: list.query,
    items: list.items,
    pageInfo: list.pageInfo,
    loading: list.loading,
    error,
    message,
    updateQuery: list.updateQuery,
    reload: list.reload,
    currentCount,
    otherCount,
    revoke,
  }
}

export type SessionsViewModel = ReturnType<typeof createSessionsViewModel>
