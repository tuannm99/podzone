import { Button } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { createFormStore, jsonArray, required } from '@/solid/forms'
import type { RoleBoundaryFormValues, TrustPolicyFormValues } from './forms'
import { SimulationInputsSection } from './SimulationInputsSection'
import { SimulationResultSection } from './SimulationResultSection'
import { TrustPolicySection } from './TrustPolicySection'
import { useAdminIamTrustSim } from './context'

export function TrustSimulationPanel() {
    const trust = useAdminIamTrustSim()
    const trustPolicyForm = createFormStore<TrustPolicyFormValues>({
        initialValues: {
            roleName: trust.trustRoleName(),
            trustJson: trust.trustJson(),
        },
        validators: {
            roleName: [required('Enter a role name.')],
            trustJson: [jsonArray('Trust policy must be a JSON array.')],
        },
    })
    const roleBoundaryForm = createFormStore<RoleBoundaryFormValues>({
        initialValues: {
            policyName: trust.trustBoundaryPolicyName(),
        },
        validators: {
            policyName: [required('Choose a role boundary policy.')],
        },
    })

    const saveTrustPolicy = async () => {
        if (!trustPolicyForm.validate()) return
        trustPolicyForm.setSubmitting(true)
        try {
            trust.setTrustRoleName(trustPolicyForm.values.roleName)
            trust.setTrustJson(trustPolicyForm.values.trustJson)
            await trust.handleSaveTrustPolicy()
        } finally {
            trustPolicyForm.setSubmitting(false)
        }
    }

    const saveRoleBoundary = async () => {
        if (!roleBoundaryForm.validate()) return
        roleBoundaryForm.setSubmitting(true)
        try {
            trust.setTrustBoundaryPolicyName(roleBoundaryForm.values.policyName)
            await trust.handleSaveRoleBoundary()
        } finally {
            roleBoundaryForm.setSubmitting(false)
        }
    }

    return (
        <>
            <SectionTitle
                title="Role trust and access simulation"
                subtitle="Edit trust policies, test service principals, session tags, conditions, boundaries, and SCP outcomes."
            />
            <TrustPolicySection
                trustPolicyForm={trustPolicyForm}
                roleBoundaryForm={roleBoundaryForm}
                onSaveTrustPolicy={saveTrustPolicy}
                onSaveRoleBoundary={saveRoleBoundary}
            />
            <SimulationInputsSection />
            <Button size="sm" color="dark" onClick={trust.handleSimulate}>
                Simulate access
            </Button>
            <SimulationResultSection />
        </>
    )
}
