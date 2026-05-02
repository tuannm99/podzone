type SetupDraft = {
  id: string;
  name: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  status: string;
  notes: string;
};

type CandidateVariant = {
  id: string;
  label: string;
  color: string;
  size: string;
  status: string;
};

type ArtworkChecklist = {
  frontArtwork: boolean;
  backArtwork: boolean;
  mockupReady: boolean;
  printSpecChecked: boolean;
};

type CatalogCandidate = {
  id: string;
  draftId: string;
  title: string;
  sku: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  estimatedMargin: string;
  status: string;
  channel: string;
  updatedAt: string;
  variants: CandidateVariant[];
  artworkChecklist: ArtworkChecklist;
  merchandisingNotes: string;
};

type MockOrder = {
  id: string;
  candidateId: string;
  productTitle: string;
  partner: string;
  quantity: number;
  total: string;
  customerName: string;
  status: string;
  timeline: string[];
  exceptionType: string;
  exceptionStatus: string;
};

export function productSetupStorageKey(tenantId: string) {
  return `podzone:product-setup:${tenantId}`;
}

export function ordersStorageKey(tenantId: string) {
  return `podzone:mock-orders:${tenantId}`;
}

export function seedDemoStoreState(tenantId: string) {
  const now = new Date();
  const publishedAt = now.toISOString();
  const readyAt = new Date(now.getTime() - 15 * 60 * 1000).toISOString();

  const drafts: SetupDraft[] = [
    {
      id: 'draft-tee-seeded',
      name: 'Signature Tee',
      partner: 'Acme Print Lab',
      baseCost: '$8.20',
      retailPrice: '$24.00',
      status: 'ready_for_review',
      notes: 'Front print approved and mockups nearly ready.',
    },
    {
      id: 'draft-poster-seeded',
      name: 'Studio Poster',
      partner: 'North Fulfillment',
      baseCost: '$6.40',
      retailPrice: '$19.00',
      status: 'draft',
      notes: 'Need final paper stock confirmation before publish.',
    },
  ];

  const candidates: CatalogCandidate[] = [
    {
      id: 'candidate-tee-seeded',
      draftId: 'draft-tee-seeded',
      title: 'Signature Tee',
      sku: 'signature-tee-1001',
      partner: 'Acme Print Lab',
      baseCost: '$8.20',
      retailPrice: '$24.00',
      estimatedMargin: '$15.80',
      status: 'published_mock',
      channel: 'website_store',
      updatedAt: publishedAt,
      variants: [
        {
          id: 'variant-tee-black-m',
          label: 'Black / M',
          color: 'Black',
          size: 'M',
          status: 'ready',
        },
        {
          id: 'variant-tee-white-l',
          label: 'White / L',
          color: 'White',
          size: 'L',
          status: 'ready',
        },
      ],
      artworkChecklist: {
        frontArtwork: true,
        backArtwork: false,
        mockupReady: true,
        printSpecChecked: true,
      },
      merchandisingNotes: 'Best launch product for creator-led drops.',
    },
    {
      id: 'candidate-poster-seeded',
      draftId: 'draft-poster-seeded',
      title: 'Studio Poster',
      sku: 'studio-poster-1002',
      partner: 'North Fulfillment',
      baseCost: '$6.40',
      retailPrice: '$19.00',
      estimatedMargin: '$12.60',
      status: 'ready',
      channel: 'marketplace_mock',
      updatedAt: readyAt,
      variants: [
        {
          id: 'variant-poster-a3',
          label: 'Matte / A3',
          color: 'Matte',
          size: 'A3',
          status: 'ready',
        },
      ],
      artworkChecklist: {
        frontArtwork: true,
        backArtwork: false,
        mockupReady: true,
        printSpecChecked: false,
      },
      merchandisingNotes: 'Ready for second wave once print spec is signed off.',
    },
  ];

  const orders: MockOrder[] = [
    {
      id: 'ORD-200101',
      candidateId: 'candidate-tee-seeded',
      productTitle: 'Signature Tee',
      partner: 'Acme Print Lab',
      quantity: 2,
      total: '$48.00',
      customerName: 'Nguyen Minh',
      status: 'in_production',
      timeline: [
        'Order created for Signature Tee',
        'Queued for Acme Print Lab',
        'Sent to Acme Print Lab for POD production',
      ],
      exceptionType: '',
      exceptionStatus: '',
    },
    {
      id: 'ORD-200102',
      candidateId: 'candidate-tee-seeded',
      productTitle: 'Signature Tee',
      partner: 'Acme Print Lab',
      quantity: 1,
      total: '$24.00',
      customerName: 'Tran Mai',
      status: 'queued',
      timeline: [
        'Order created for Signature Tee',
        'Queued for Acme Print Lab',
        'Exception opened: partner delay',
      ],
      exceptionType: 'partner_delay',
      exceptionStatus: 'open',
    },
    {
      id: 'ORD-200103',
      candidateId: 'candidate-tee-seeded',
      productTitle: 'Signature Tee',
      partner: 'Acme Print Lab',
      quantity: 1,
      total: '$24.00',
      customerName: 'Le An',
      status: 'shipped',
      timeline: [
        'Order created for Signature Tee',
        'Queued for Acme Print Lab',
        'Sent to Acme Print Lab for POD production',
        'Marked as shipped from Acme Print Lab',
      ],
      exceptionType: '',
      exceptionStatus: '',
    },
  ];

  localStorage.setItem(
    productSetupStorageKey(tenantId),
    JSON.stringify({ drafts, candidates })
  );
  localStorage.setItem(ordersStorageKey(tenantId), JSON.stringify(orders));
}

export function resetDemoStoreState(tenantId: string) {
  localStorage.removeItem(productSetupStorageKey(tenantId));
  localStorage.removeItem(ordersStorageKey(tenantId));
}
