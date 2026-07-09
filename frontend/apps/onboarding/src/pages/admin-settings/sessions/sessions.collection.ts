import type { CollectionFilterField } from '@podzone/shared/ui/components/common/CollectionFilters'
import type { CollectionSortOption } from '@podzone/shared/ui/components/common/CollectionToolbar'

export const sessionSortOptions: CollectionSortOption[] = [
    { label: 'Created time', value: 'created_at' },
    { label: 'Expiry time', value: 'expires_at' },
    { label: 'Status', value: 'status' },
    { label: 'Workspace', value: 'tenant_id' },
]

export const sessionFilterFields: CollectionFilterField[] = [
    {
        label: 'Session id',
        value: 'id',
        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
    },
    {
        label: 'Workspace',
        value: 'active_tenant_id',
        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS', 'FILTER_OPERATOR_IN'],
    },
    {
        label: 'Status',
        value: 'status',
        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_NEQ', 'FILTER_OPERATOR_IN'],
    },
    {
        label: 'Created time',
        value: 'created_at',
        operators: ['FILTER_OPERATOR_GT', 'FILTER_OPERATOR_GTE', 'FILTER_OPERATOR_LT', 'FILTER_OPERATOR_LTE'],
    },
]
