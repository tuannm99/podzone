import { useParams } from '@tanstack/solid-router';
import { createEffect, createSignal } from 'solid-js';
import { TENANT_GQL_URL } from '../../../services/baseurl';
import {
  ordersStorageKey,
  resetDemoStoreState,
  seedDemoStoreState,
} from '../../../services/demoStore';
import { getProductSetupSnapshot } from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import { PageShell } from '../../components/common/PageShell';
import { EmptyBlock, ErrorAlert, InfoAlert } from '../../components/common/Feedback';
import { Badge, Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

type MockOrder = {
  id: string;
  status: string;
  exceptionStatus: string;
  total: string;
  partner: string;
};

function parseMoney(value: string) {
  const cleaned = value.replace(/[^0-9.]/g, '');
  const parsed = Number.parseFloat(cleaned);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
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
  const [estimatedMarginTotal, setEstimatedMarginTotal] = createSignal('$0.00');
  const [topPartnerLoad, setTopPartnerLoad] = createSignal('No partner load yet');
  const [issueRate, setIssueRate] = createSignal('0%');
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const loadOpsSnapshot = async () => {
    setError('');
    const productResult = await getProductSetupSnapshot();
    if (productResult.success) {
      setDraftCount(productResult.data.drafts.length);
      const publishedCandidates = productResult.data.candidates.filter(
          (candidate) => candidate.status === 'published_mock'
      );
      setPublishedCandidateCount(publishedCandidates.length);
      setEstimatedMarginTotal(
        formatMoney(
          publishedCandidates.reduce(
            (sum, candidate) => sum + parseMoney(candidate.estimatedMargin),
            0
          )
        )
      );
    } else {
      setError(productResult.message);
      setDraftCount(0);
      setPublishedCandidateCount(0);
      setEstimatedMarginTotal('$0.00');
    }

    const ordersRaw = localStorage.getItem(ordersStorageKey(params().tenantId));
    if (ordersRaw) {
      try {
        const parsed = JSON.parse(ordersRaw) as MockOrder[];
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
        const loadByPartner = parsed.reduce<Record<string, number>>((acc, order) => {
          acc[order.partner] = (acc[order.partner] || 0) + 1;
          return acc;
        }, {});
        const topPartner = Object.entries(loadByPartner).sort((a, b) => b[1] - a[1])[0];
        setTopPartnerLoad(
          topPartner ? `${topPartner[0]} · ${topPartner[1]} orders` : 'No partner load yet'
        );
        const activeIssues = parsed.filter(
          (order) =>
            order.exceptionStatus === 'open' ||
            order.exceptionStatus === 'escalated'
        ).length;
        setIssueRate(
          parsed.length > 0 ? `${Math.round((activeIssues / parsed.length) * 100)}%` : '0%'
        );
      } catch {
        setInProductionCount(0);
        setOpenExceptionCount(0);
        setMockRevenue('$0.00');
        setTopPartnerLoad('No partner load yet');
        setIssueRate('0%');
      }
    } else {
      setInProductionCount(0);
      setOpenExceptionCount(0);
      setMockRevenue('$0.00');
      setTopPartnerLoad('No partner load yet');
      setIssueRate('0%');
    }
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    setTenantReady(tokenStorage.getActiveTenantID() === params().tenantId);
    void loadOpsSnapshot();
  });

  const seedDemo = () => {
    seedDemoStoreState(params().tenantId);
    loadOpsSnapshot();
    setMessage('Seeded local demo data for this store. Print partners remain managed separately in the partner service.');
  };

  const resetDemo = () => {
    resetDemoStoreState(params().tenantId);
    loadOpsSnapshot();
    setMessage('Cleared local demo data for this store. Partner records in the backend were left untouched.');
  };

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Store Workspace Prototype"
          title={`Store ${params().tenantId}`}
          copy="This store workspace stays scoped to the active session. Product setup is now backend-backed for the store, while order flow and some dashboard metrics still run as prototype data."
        />
      </Card>

      {message() ? <InfoAlert>{message()}</InfoAlert> : null}

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
        <StatCard label="Estimated margin pool" value={estimatedMarginTotal()} />
        <StatCard label="Top partner load" value={topPartnerLoad()} />
        <StatCard label="Issue rate" value={issueRate()} />
      </div>

      <Card class="space-y-4">
        <SectionTitle
          title="Prototype controls"
          subtitle="Seed local-only sample orders to exercise the POD workflow concept without depending on backend order or fulfillment services yet."
        />
        <div class="flex flex-wrap gap-3">
          <Button color="green" onClick={seedDemo}>
            Seed prototype data
          </Button>
          <Button color="light" onClick={resetDemo}>
            Reset prototype data
          </Button>
        </div>
        <p class="text-sm text-gray-600">
          These actions now affect browser-local order prototype data only. Product setup drafts and candidates remain persisted in the backend.
        </p>
      </Card>

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
          title="Prototype operations pulse"
          subtitle="Product setup metrics come from the backend. Order and exception metrics still come from local prototype data."
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
          <Badge
            content={`revenue ${mockRevenue()}`}
            color="indigo"
          />
          <Badge
            content={`issue rate ${issueRate()}`}
            color={openExceptionCount() > 0 ? 'yellow' : 'green'}
          />
        </div>
        {!publishedCandidateCount() ? (
          <EmptyBlock
            title="No published products yet"
            copy="Start in Product setup, promote a candidate, and prototype publish it before testing the rest of this local workflow."
          />
        ) : null}
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Start here"
          subtitle="A simple guided prototype flow that separates the real partner record layer from local-only product and order experiments."
        />
        <div class="grid gap-4 md:grid-cols-3">
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
              Route, monitor, and resolve
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Use mock orders to test production flow, shipping status, and exception handling.
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
          subtitle="A lightweight view of how this experimental POD store is performing before any real finance or analytics stack exists."
        />
        <div class="grid gap-4 md:grid-cols-2">
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Margin outlook
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {estimatedMarginTotal()}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Based on the sum of estimated margin from backend-backed published mock products.
            </p>
          </div>
          <div class="rounded-2xl border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
              Partner pressure
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900">
              {topPartnerLoad()}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              Highlights which print partner is carrying most of the current mock routing load.
            </p>
          </div>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Action shortcuts"
          subtitle="Jump straight to the next likely action based on the current experimental state."
        />
        <div class="flex flex-wrap gap-3">
          <Button href={`/t/${params().tenantId}/products/setup`} color="green">
            {publishedCandidateCount() > 0 ? 'Refine product candidates' : 'Publish first mock product'}
          </Button>
          <Button href={`/t/${params().tenantId}/orders`} color="blue">
            {openExceptionCount() > 0 ? 'Review active issues' : 'Review routing board'}
          </Button>
          <Button href={`/t/${params().tenantId}/orders`} color="alternative">
            {inProductionCount() > 0 ? 'Track orders in production' : 'Create first routed order'}
          </Button>
          <Button href={`/t/${params().tenantId}/partners`} color="light">
            {topPartnerLoad() === 'No partner load yet' ? 'Set up print partners' : 'Review partner load'}
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
