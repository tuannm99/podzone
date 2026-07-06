export type CreatePolicyFormValues = {
    scope: string
    name: string
    description: string
    statementsJson: string
}

export type CreatePolicyVersionFormValues = {
    statementsJson: string
}
