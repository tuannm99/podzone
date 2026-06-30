import { Button } from '@/solid/components/common/Primitives'
import { useAdminIamTrustSim } from './context'

export function SimulationPresetButtons() {
  const trust = useAdminIamTrustSim()
  return (
    <div class="flex flex-wrap gap-2">
      <Button size="xs" color="light" onClick={trust.applyServiceAssumePreset}>
        Preset: service assume
      </Button>
      <Button size="xs" color="light" onClick={trust.applyTenantAssumePreset}>
        Preset: tenant admin assume
      </Button>
      <Button size="xs" color="light" onClick={trust.applyScopeDownDenyPreset}>
        Preset: scope-down deny
      </Button>
    </div>
  )
}
