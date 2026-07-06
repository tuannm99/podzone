import type { CollectionFilterField } from '@/solid/components/common/CollectionFilters'
import type { CollectionSortOption } from '@/solid/components/common/CollectionToolbar'

export const auditSortOptions: CollectionSortOption[] = [
    { label: 'Created time', value: 'created_at' },
    { label: 'Action', value: 'action' },
    { label: 'Resource type', value: 'resource_type' },
    { label: 'Status', value: 'status' },
    { label: 'Workspace', value: 'tenant_id' },
]

export const auditFilterFields: CollectionFilterField[] = [
    {
        label: 'Action',
        value: 'action',
        operators: [
            'FILTER_OPERATOR_EQ',
            'FILTER_OPERATOR_CONTAINS',
            'FILTER_OPERATOR_STARTS_WITH',
            'FILTER_OPERATOR_IN',
        ],
    },
    {
        label: 'Resource type',
        value: 'resource_type',
        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
    },
    {
        label: 'Resource id',
        value: 'resource_id',
        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
    },
    {
        label: 'Workspace',
        value: 'tenant_id',
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
