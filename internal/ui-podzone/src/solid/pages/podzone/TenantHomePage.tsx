import { useParams } from '@tanstack/solid-router';
import { createEffect, createSignal } from 'solid-js';
import { TENANT_GQL_URL } from '../../../services/baseurl';
import { getRoutedOrders } from '../../../services/orders';
import { getProductSetupSnapshot } from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import { PageShell } from '../../components/common/PageShell';
import { EmptyBlock, ErrorAlert } from '../../components/common/Feedback';
import { Badge, Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

function parseMoney(value: string) {
  const cleaned = value.replace(/[^0-9.]/g, '');
  const parsed = Number.parseFloat(cleaned);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

function isOverdue(value?: string) {
  if (!value) {
    return false;
  }
  return new Date(value).getTime() < Date.now();
}

function buildOrdersHref(
  tenantID: string,
  queueView: string,
  queueSort: string = 'priority',
  operatorLens?: string
) {
  const params = new URLSearchParams({
    queueView,
    queueSort,
  });
  if (operatorLens) {
    params.set('operatorLens', operatorLens);
  }
  return `/t/${tenantID}/orders?${params.toString()}`;
}

export default function TenantHomePage() {
  const params = useParams({ from: '/t/$tenantId' });
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
    void loadOpsSnapshot();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Store Workspace"
          title={`Store ${params().tenantId}`}
          copy="This store workspace stays scoped to the active session. Product setup, order routing, shipment control, and settlement readiness now come from backend store data."
        />
      </Card>

      {error() ? <ErrorAlert>{error()}</ErrorAlert> : null}

      <div class="grid gap-4 md:grid-cols-3">
        <StatCard label="Store id" value={params().tenantId} />
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

      <Card class="space-y-4">
        <SectionTitle
          title="Current store context"
          subtitle="Requests in this workspace rely on the active store in the signed-in session. The local route value is only used for navigation."
        />
        <div class="flex flex-wrap gap-2">
          <Badge
            content={`current store: ${tokenStorage.getActiveTenantID() || 'missing'}`}
            color={tenantReady() ? 'green' : 'yellow'}
          />
          <Badge
            content={`route store: ${tenantStorage.getTenantID() || params().tenantId}`}
            color="indigo"
          />
          <Badge content="Authorization: Bearer ..." color="green" />
        </div>
        {!tenantReady() ? (
          <EmptyBlock
            title="Store session not ready"
            copy="The client could not confirm this store as the current active workspace yet."
          />
        ) : null}
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="POD operations pulse"
          subtitle="Product setup, order routing, shipment control, and settlement metrics now come from backend store data."
        />
        <div class="flex flex-wrap gap-2">
          <Badge
            content={`${publishedCandidateCount()} published mock products`}
            color={publishedCandidateCount() > 0 ? 'green' : 'dark'}
          />
          <Badge
            content={`${inProductionCount()} orders in production`}
            color={inProductionCount() > 0 ? 'blue' : 'dark'}
          />
          <Badge
            content={`${openExceptionCount()} active issues`}
            color={openExceptionCount() > 0 ? 'yellow' : 'green'}
          />
          <Badge content={`revenue ${mockRevenue()}`} color="indigo" />
          <Badge content={`margin ${realizedMarginTotal()}`} color="blue" />
          <Badge content={`issue cost ${issueCostExposure()}`} color="red" />
          <Badge
            content={`${shipmentSlaOverdueCount()} shipment SLA overdue`}
            color={shipmentSlaOverdueCount() > 0 ? 'red' : 'green'}
          />
          <Badge
            content={`${issueSlaOverdueCount()} issue SLA overdue`}
            color={issueSlaOverdueCount() > 0 ? 'red' : 'green'}
          />
          <Badge
            content={`${pendingSettlementCount()} pending settlements`}
            color={pendingSettlementCount() > 0 ? 'yellow' : 'green'}
          />
          <Badge
            content={`issue rate ${issueRate()}`}
            color={openExceptionCount() > 0 ? 'yellow' : 'green'}
          />
        </div>
        {!publishedCandidateCount() ? (
          <EmptyBlock
            title="No published products yet"
            copy="Start in Product setup, promote a candidate, and mock publish it before testing the rest of this POD workflow."
          />
        ) : null}
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Start here"
          subtitle="A simple guided POD flow that separates the partner record layer from store-side catalog and routing operations."
        />
        <div class="grid gap-4 md:grid-cols-4">
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              1. Partner setup
            </p>
            <p class="mt-2 text-base font-semibold text-gray-900">
              Confirm who prints and fulfills
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Keep partner records current before shaping products or routing test orders.
            </p>
            <div class="mt-4">
              <Button href={`/t/${params().tenantId}/partners`} color="blue">
                Review print partners
              </Button>
            </div>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              2. Product setup
            </p>
            <p class="mt-2 text-base font-semibold text-gray-900">
              Build candidates for the store catalog
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Create backend-backed drafts, verify artwork readiness, and mock publish what is ready to route.
            </p>
            <div class="mt-4">
              <Button href={`/t/${params().tenantId}/products/setup`} color="green">
                Open product setup
              </Button>
            </div>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              3. Order operations
            </p>
            <p class="mt-2 text-base font-semibold text-gray-900">
              Route, ship, and settle
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Route backend-backed store orders through production flow, manual shipment status, and settlement readiness.
            </p>
            <div class="mt-4">
              <Button href={`/t/${params().tenantId}/orders`} color="alternative">
                Open orders board
              </Button>
            </div>
          </div>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Commercial snapshot"
          subtitle="A lightweight operational finance view from backend-backed routed orders before any separate analytics stack exists."
        />
        <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Realized margin
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {realizedMarginTotal()}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Calculated from routed order revenue minus fulfillment and shipping costs captured in the store workflow.
            </p>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Settlement pressure
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {pendingSettlementCount()} pending · {disputedSettlementCount()} disputed
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Highlights which routed orders still need reconciliation or manual finance follow-up.
            </p>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Issue cost exposure
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {issueCostExposure()}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Captures reprint, delivery issue, and other exception costs now reducing realized margin.
            </p>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Queue pressure
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {shipmentSlaOverdueCount()} shipment · {issueSlaOverdueCount()} issue
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Tracks overdue shipment and issue follow-up deadlines on the operator queue.
            </p>
          </div>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Action shortcuts"
          subtitle="Jump straight to the next likely action based on the current operational state."
        />
        <div class="flex flex-wrap gap-3">
          <Button href={`/t/${params().tenantId}/products/setup`} color="green">
            {publishedCandidateCount() > 0 ? 'Refine product candidates' : 'Publish first mock product'}
          </Button>
          <Button href={`/t/${params().tenantId}/orders`} color="blue">
            {openExceptionCount() > 0 ? 'Review active issues' : 'Review routing board'}
          </Button>
          <Button href={`/t/${params().tenantId}/orders/finance`} color="alternative">
            {pendingSettlementCount() > 0 ? 'Review settlement queue' : 'Create first routed order'}
          </Button>
          <Button href={`/t/${params().tenantId}/partners`} color="light">
            {topPartnerLoad() === 'No partner load yet' ? 'Set up print partners' : 'Review partner load'}
          </Button>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Queue Shortcuts"
          subtitle="Open the orders board already focused on the slice that needs attention."
        />
        <div class="flex flex-wrap gap-3">
          <Button
            href={buildOrdersHref(params().tenantId, 'overdue')}
            color={
              shipmentSlaOverdueCount() > 0 || issueSlaOverdueCount() > 0
                ? 'red'
                : 'alternative'
            }
          >
            Overdue queue
          </Button>
          <Button
            href={buildOrdersHref(params().tenantId, 'delivery_issues')}
            color={openExceptionCount() > 0 ? 'dark' : 'alternative'}
          >
            Issue queue
          </Button>
          <Button
            href={buildOrdersHref(params().tenantId, 'settlement_pending')}
            color={pendingSettlementCount() > 0 ? 'green' : 'alternative'}
          >
            Settlement follow-up
          </Button>
          <Button
            href={`/t/${params().tenantId}/orders/finance`}
            color={disputedSettlementCount() > 0 ? 'red' : 'alternative'}
          >
            Finance review
          </Button>
          <Button
            href={buildOrdersHref(params().tenantId, 'all', 'priority')}
            color="blue"
          >
            Priority queue
          </Button>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Operations"
          subtitle="Direct links into each experimental POD workflow area."
        />
        <div class="flex flex-wrap gap-3">
          <Button href={`/t/${params().tenantId}/products/setup`} color="green">
            Open product setup
          </Button>
          <Button href={`/t/${params().tenantId}/partners`} color="blue">
            Open print partners
          </Button>
          <Button href={`/t/${params().tenantId}/orders`} color="alternative">
            Open orders
          </Button>
        </div>
      </Card>
    </PageShell>
  );
}
