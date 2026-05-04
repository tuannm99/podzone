import { BACKOFFICE_API_URL } from './baseurl';
import { http, type HttpError } from './http';

export type SetupDraft = {
  id: string;
  name: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  status: string;
  notes: string;
  createdAt?: string;
  updatedAt?: string;
};

export type CandidateVariant = {
  id: string;
  label: string;
  color: string;
  size: string;
  status: string;
};

export type ArtworkChecklist = {
  frontArtwork: boolean;
  backArtwork: boolean;
  mockupReady: boolean;
  printSpecChecked: boolean;
};

export type CatalogCandidate = {
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

export type ProductSetupSnapshot = {
  drafts: SetupDraft[];
  candidates: CatalogCandidate[];
};

type ProductSetupResult<T> =
  | { success: true; data: T }
  | { success: false; message: string };

type CreateDraftPayload = {
  name: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  status: string;
  notes: string;
};

type PromoteCandidatePayload = {
  draftId: string;
  channel: string;
  variantColor: string;
  variantSize: string;
  artworkChecklist: ArtworkChecklist;
  merchandisingNotes: string;
};

function toFailure(error: unknown, fallback: string): ProductSetupResult<never> {
  const message =
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
      ? error.message
      : fallback;
  return { success: false, message };
}

export async function getProductSetupSnapshot(): Promise<
  ProductSetupResult<ProductSetupSnapshot>
> {
  try {
    const { data } = await http.get<ProductSetupSnapshot>(
      '/backoffice/v1/product-setup',
      { baseURL: BACKOFFICE_API_URL }
    );
    return {
      success: true,
      data: {
        drafts: data.drafts || [],
        candidates: data.candidates || [],
      },
    };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load product setup');
  }
}

export async function createProductSetupDraft(
  payload: CreateDraftPayload
): Promise<ProductSetupResult<SetupDraft>> {
  try {
    const { data } = await http.post<SetupDraft>(
      '/backoffice/v1/product-setup/drafts',
      payload,
      { baseURL: BACKOFFICE_API_URL }
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create setup draft');
  }
}

export async function promoteProductSetupCandidate(
  payload: PromoteCandidatePayload
): Promise<ProductSetupResult<CatalogCandidate>> {
  try {
    const { data } = await http.post<CatalogCandidate>(
      '/backoffice/v1/product-setup/candidates/promote',
      payload,
      { baseURL: BACKOFFICE_API_URL }
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to promote setup candidate');
  }
}

export async function updateProductSetupCandidateStatus(
  id: string,
  status: string
): Promise<ProductSetupResult<CatalogCandidate>> {
  try {
    const { data } = await http.patch<CatalogCandidate>(
      `/backoffice/v1/product-setup/candidates/${id}/status`,
      { status },
      { baseURL: BACKOFFICE_API_URL }
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to update setup candidate status'
    );
  }
}
