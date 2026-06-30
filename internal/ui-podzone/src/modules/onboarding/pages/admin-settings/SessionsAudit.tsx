import { AuditPanel } from './audit/AuditPanel'
import { SessionsPanel } from './sessions/SessionsPanel'

export function SessionsAudit() {
  return (
    <>
      <SessionsPanel />
      <AuditPanel />
    </>
  )
}
