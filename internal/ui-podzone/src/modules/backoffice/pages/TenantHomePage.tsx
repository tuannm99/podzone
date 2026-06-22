import { useParams } from '@tanstack/solid-router';
import { createEffect, createSignal } from 'solid-js';
import { TENANT_GQL_URL } from '@/services/baseurl';
import { getRoutedOrders } from '@/services/orders';
import { getProductSetupSnapshot } from '@/services/productSetup';
import { tenantStorage } from '@/services/tenantStorage';
import { tokenStorage } from '@/services/tokenStorage';
import { PageShell } from '@/solid/components/common/PageShell';
import { EmptyBlock, ErrorAlert } from '@/solid/components/common/Feedback';
import { Card } from '@/solid/components/common/Primitives';
import { SectionLead } from '@/solid/components/common/SectionLead';
import { StatCard } from '@/solid/components/dashboard/StatCard';
import { useTenantWorkspace } from '@/solid/workspace/context';
import { TenantHomeSections } from './home/TenantHomeSections';
import { formatMoney, isOverdue, parseMoney } from './home/presentation';

export default function TenantHomePage() {
  const params = useParams({ from: '/t/$tenantId' });
  const workspace = useTenantWorkspace();
  const [tenantReady, setTenantReady] = createSignal(
    tokenStorage.getActiveTenantID() === params().tenantId
  );
  const [draftCount, setDraftCount] = createSignal(0);
  const [publishedCandidateCount, setPublishedCandidateCount] = createSignal(0);
  const [inProductionCount, setInProductionCount] = createSignal(0);
  const [openExceptionCount, setOpenExceptionCount] = createSignal(0);
  const [mockRevenue, setMockRevenue] = createSignal('$0.00');
  const [realizedMarginTotal, setRealizedMarginTotal] = createSignal('$0.00');
  const [pendingSettlementCount, setPendingSettlementCount] = createSignal(0);
  const [disputedSettlementCount, setDisputedSettlementCount] = createSignal(0);
  const [issueCostExposure, setIssueCostExposure] = createSignal('$0.00');
  const [shipmentSlaOverdueCount, setShipmentSlaOverdueCount] = createSignal(0);
  const [issueSlaOverdueCount, setIssueSlaOverdueCount] = createSignal(0);
  const [topPartnerLoad, setTopPartnerLoad] = createSignal('No partner load yet');
  const [issueRate, setIssueRate] = createSignal('0%');
  const [error, setError] = createSignal('');
  const currentStoreId = () => workspace?.currentStoreId() || '';
  const currentStore = () => workspace?.currentStore();
  const workspaceReady = () => !workspace || currentStoreId().trim().length > 0;
  const storeLabel = () =>
    currentStore()?.name || currentStoreId() || 'Select a store';

  const loadOpsSnapshot = async () => {
    setError('');

    const productResult = await getProductSetupSnapshot();
    if (productResult.success) {
      setDraftCount(productResult.data.drafts.length);
      setPublishedCandidateCount(
        productResult.data.candidates.filter(
          (candidate) => candidate.status === 'published_mock'
        ).length
      );
    } else {
      setError(productResult.message);
      setDraftCount(0);
      setPublishedCandidateCount(0);
    }

    const orderResult = await getRoutedOrders();
    if (orderResult.success) {
      const parsed = orderResult.data.orders;
      setInProductionCount(
        parsed.filter((order) => order.status === 'in_production').length
      );
      setOpenExceptionCount(
        parsed.filter(
          (order) =>
            order.exceptionStatus === 'open' ||
            order.exceptionStatus === 'escalated'
        ).length
      );
      setMockRevenue(
        formatMoney(
          parsed.reduce((sum, order) => sum + parseMoney(order.total), 0)
        )
      );
      setRealizedMarginTotal(
        formatMoney(
          parsed.reduce(
            (sum, order) => sum + parseMoney(order.realizedMargin),
            0
          )
        )
      );
      setPendingSettlementCount(
        parsed.filter((order) => order.settlementStatus === 'pending').length
      );
      setDisputedSettlementCount(
        parsed.filter((order) => order.settlementStatus === 'disputed').length
      );
      setIssueCostExposure(
        formatMoney(
          parsed.reduce((sum, order) => sum + parseMoney(order.issueCost), 0)
        )
      );
      setShipmentSlaOverdueCount(
        parsed.filter(
          (order) =>
            !!order.shipmentSlaDueAt &&
            isOverdue(order.shipmentSlaDueAt) &&
            order.shipmentStatus !== 'delivered'
        ).length
      );
      setIssueSlaOverdueCount(
        parsed.filter(
          (order) =>
            !!order.issueSlaDueAt &&
            isOverdue(order.issueSlaDueAt) &&
            (order.exceptionStatus === 'open' ||
              order.exceptionStatus === 'escalated' ||
              order.shipmentStatus === 'delivery_issue')
        ).length
      );
      const loadByPartner = parsed.reduce<Record<string, number>>((acc, order) => {
        acc[order.partner] = (acc[order.partner] || 0) + 1;
        return acc;
      }, {});
      const topPartner = Object.entries(loadByPartner).sort(
        (a, b) => b[1] - a[1]
      )[0];
      setTopPartnerLoad(
        topPartner ? `${topPartner[0]} · ${topPartner[1]} orders` : 'No partner load yet'
      );
      const activeIssues = parsed.filter(
        (order) =>
          order.exceptionStatus === 'open' ||
          order.exceptionStatus === 'escalated'
      ).length;
      setIssueRate(
        parsed.length > 0
          ? `${Math.round((activeIssues / parsed.length) * 100)}%`
          : '0%'
      );
    } else {
      setError((current) => current || orderResult.message);
      setInProductionCount(0);
      setOpenExceptionCount(0);
      setMockRevenue('$0.00');
      setRealizedMarginTotal('$0.00');
      setPendingSettlementCount(0);
      setDisputedSettlementCount(0);
      setIssueCostExposure('$0.00');
      setShipmentSlaOverdueCount(0);
      setIssueSlaOverdueCount(0);
      setTopPartnerLoad('No partner load yet');
      setIssueRate('0%');
    }
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    setTenantReady(tokenStorage.getActiveTenantID() === params().tenantId);
    if (!workspaceReady()) {
      return;
    }
    void loadOpsSnapshot();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Store Workspace"
          title={storeLabel()}
          copy="This workspace stays scoped to the active tenant session and the selected store. Product setup, order routing, shipment control, and settlement readiness now resolve against backend store data."
        />
      </Card>

      {error() ? <ErrorAlert>{error()}</ErrorAlert> : null}
      {!workspaceReady() ? (
        <EmptyBlock
          title="Choose a store first"
          copy="Use the store switcher in the workspace shell before loading store-scoped POD operations."
        />
      ) : null}

      <div class="grid gap-4 md:grid-cols-3">
        <StatCard label="Store id" value={currentStoreId() || 'pending'} />
        <StatCard label="Transport" value="GraphQL" />
        <StatCard
          label="Endpoint"
          value={TENANT_GQL_URL.replace(/^https?:\/\//, '')}
        />
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Setup drafts" value={String(draftCount())} />
        <StatCard
          label="Published mock products"
          value={String(publishedCandidateCount())}
        />
        <StatCard
          label="Orders in production"
          value={String(inProductionCount())}
        />
        <StatCard
          label="Open exceptions"
          value={String(openExceptionCount())}
        />
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Mock revenue" value={mockRevenue()} />
        <StatCard label="Realized margin" value={realizedMarginTotal()} />
        <StatCard
          label="Pending settlements"
          value={String(pendingSettlementCount())}
        />
        <StatCard label="Issue cost" value={issueCostExposure()} />
        <StatCard
          label="Shipment SLA overdue"
          value={String(shipmentSlaOverdueCount())}
        />
        <StatCard label="Top partner load" value={topPartnerLoad()} />
      </div>

      <TenantHomeSections
        tenantId={params().tenantId}
        currentStoreId={currentStoreId}
        tenantReady={tenantReady}
        publishedCandidateCount={publishedCandidateCount}
        inProductionCount={inProductionCount}
        openExceptionCount={openExceptionCount}
        mockRevenue={mockRevenue}
        realizedMarginTotal={realizedMarginTotal}
        pendingSettlementCount={pendingSettlementCount}
        disputedSettlementCount={disputedSettlementCount}
        issueCostExposure={issueCostExposure}
        shipmentSlaOverdueCount={shipmentSlaOverdueCount}
        issueSlaOverdueCount={issueSlaOverdueCount}
        topPartnerLoad={topPartnerLoad}
        issueRate={issueRate}
      />
    </PageShell>
  );
}
