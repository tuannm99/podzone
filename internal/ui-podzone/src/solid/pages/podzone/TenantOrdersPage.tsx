import { useParams } from '@tanstack/solid-router';
import { createEffect, For } from 'solid-js';
import { tenantStorage } from '../../../services/tenantStorage';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const mockOrders = [
  { id: 'ORD-1001', status: 'Processing', total: '$128.00' },
  { id: 'ORD-1002', status: 'Shipped', total: '$64.50' },
  { id: 'ORD-1003', status: 'Pending', total: '$19.99' },
];

export default function TenantOrdersPage() {
  const params = useParams({ from: '/t/$tenantId/orders' });

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Orders"
          title={`Orders for ${params().tenantId}`}
          copy="This page is still ready for the tenant GraphQL layer, but now uses the cloned card and list treatment from the reference UI."
        />
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Order list"
          subtitle="Placeholder data until the tenant orders query is connected."
        />
        <div class="space-y-3">
          <For each={mockOrders}>
            {(order) => (
              <div class="flex flex-col gap-3 rounded-2xl border border-gray-200 bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p class="text-base font-semibold text-gray-900">
                    {order.id}
                  </p>
                  <p class="text-sm text-gray-500">
                    tenant {params().tenantId}
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <Badge
                    content={order.status}
                    color={
                      order.status === 'Shipped'
                        ? 'green'
                        : order.status === 'Processing'
                          ? 'blue'
                          : 'yellow'
                    }
                  />
                  <span class="text-sm font-medium text-gray-700">
                    {order.total}
                  </span>
                </div>
              </div>
            )}
          </For>
        </div>
      </Card>
    </PageShell>
  );
}
