import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps, Setter } from 'solid-js'
import type { PolicyAttachmentInfo, PolicyInfo, PolicyVersionInfo } from '@podzone/shared/services/iam'
import type { CollectionQuery, PageInfo } from '@podzone/shared/services/collection'
import type { IamPermissionOption } from '../shared/IamPermissionMatrix'
import type { CreatePolicyFormValues, CreatePolicyVersionFormValues } from './forms'

export type PolicyScopeOption = {
    name: string
    value: string
}

export type PolicyOption = {
    name: string
    value: string
}

export type AdminIamPolicyContextValue = {
    permissionOptions: Accessor<IamPermissionOption[]>
    identityForUser: (userID: number | string) => {
        label: string
        description: string
    }
    policyScopeOptions: Accessor<PolicyScopeOption[]>
    policyScope: Accessor<string>
    setPolicyScope: Setter<string>
    policyName: Accessor<string>
    setPolicyName: Setter<string>
    policyDescription: Accessor<string>
    setPolicyDescription: Setter<string>
    policyStatementsJson: Accessor<string>
    setPolicyStatementsJson: Setter<string>
    policyVersionJson: Accessor<string>
    setPolicyVersionJson: Setter<string>
    selectedPolicyName: Accessor<string>
    setSelectedPolicyName: Setter<string>
    policyOptions: Accessor<PolicyOption[]>
    policies: Accessor<PolicyInfo[]>
    query: CollectionQuery
    pageInfo: Accessor<PageInfo>
    loading: Accessor<boolean>
    error: Accessor<string>
    updateQuery: (patch: Partial<CollectionQuery>) => void
    policyDetail: Accessor<PolicyInfo | undefined>
    policyVersions: Accessor<PolicyVersionInfo[]>
    policyVersionsQuery: CollectionQuery
    policyVersionsPageInfo: Accessor<PageInfo>
    policyVersionsLoading: Accessor<boolean>
    policyVersionsError: Accessor<string>
    updatePolicyVersionsQuery: (patch: Partial<CollectionQuery>) => void
    policyAttachments: Accessor<PolicyAttachmentInfo[]>
    policyAttachmentsQuery: CollectionQuery
    policyAttachmentsPageInfo: Accessor<PageInfo>
    policyAttachmentsLoading: Accessor<boolean>
    policyAttachmentsError: Accessor<string>
    updatePolicyAttachmentsQuery: (patch: Partial<CollectionQuery>) => void
    attachmentColor: (type: string) => 'blue' | 'green' | 'yellow' | 'pink' | 'dark'
    submitCreatePolicy: (event: SubmitEvent) => Promise<void>
    createPolicyFromForm: (values: CreatePolicyFormValues) => Promise<void>
    createPolicyVersionFromForm: (values: CreatePolicyVersionFormValues) => Promise<void>
    handleCreatePolicyVersion: () => Promise<void>
    handleDeletePolicy: () => Promise<void>
    handleSetDefaultVersion: (version: string) => Promise<void>
    handleDeleteVersion: (version: string) => Promise<void>
}

const AdminIamPolicyContext = createContext<AdminIamPolicyContextValue>()

export function AdminIamPolicyProvider(props: ParentProps<{ value: AdminIamPolicyContextValue }>) {
    return <AdminIamPolicyContext.Provider value={props.value}>{props.children}</AdminIamPolicyContext.Provider>
}

export function useAdminIamPolicy() {
    const ctx = useContext(AdminIamPolicyContext)
    if (!ctx) {
        throw new Error('AdminIamPolicyContext is missing')
    }
    return ctx
}
