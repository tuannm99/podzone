import { postBackofficeGraphQL } from './backofficeGraphql'

export type SetupDraft = {
  id: string
  name: string
  partner: string
  baseCost: string
  retailPrice: string
  status: string
  notes: string
  createdAt?: string
  updatedAt?: string
}

export type CandidateVariant = {
  id: string
  label: string
  color: string
  size: string
  status: string
}

export type ArtworkChecklist = {
  frontArtwork: boolean
  backArtwork: boolean
  mockupReady: boolean
  printSpecChecked: boolean
}

export type CatalogCandidate = {
  id: string
  draftId: string
  title: string
  sku: string
  partner: string
  baseCost: string
  retailPrice: string
  estimatedMargin: string
  status: string
  channel: string
  updatedAt: string
  variants: CandidateVariant[]
  artworkChecklist: ArtworkChecklist
  merchandisingNotes: string
}

export type ProductSetupSnapshot = {
  drafts: SetupDraft[]
  candidates: CatalogCandidate[]
}

type ProductSetupResult<T> =
  | { success: true; data: T }
  | { success: false; message: string }

type CreateDraftPayload = {
  name: string
  partner: string
  baseCost: string
  retailPrice: string
  status: string
  notes: string
}

type PromoteCandidatePayload = {
  draftId: string
  channel: string
  variantColor: string
  variantSize: string
  artworkChecklist: ArtworkChecklist
  merchandisingNotes: string
}

export async function getProductSetupSnapshot(): Promise<
  ProductSetupResult<ProductSetupSnapshot>
> {
  const result = await postBackofficeGraphQL<{
    productSetupSnapshot: ProductSetupSnapshot
  }>(`
    query ProductSetupSnapshot {
      productSetupSnapshot {
        drafts {
          id
          name
          partner
          baseCost
          retailPrice
          status
          notes
          createdAt
          updatedAt
        }
        candidates {
          id
          draftId
          title
          sku
          partner
          baseCost
          retailPrice
          estimatedMargin
          status
          channel
          updatedAt
          merchandisingNotes
          variants {
            id
            label
            color
            size
            status
          }
          artworkChecklist {
            frontArtwork
            backArtwork
            mockupReady
            printSpecChecked
          }
        }
      }
    }
  `)
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.productSetupSnapshot }
}

export async function createProductSetupDraft(
  payload: CreateDraftPayload
): Promise<ProductSetupResult<SetupDraft>> {
  const result = await postBackofficeGraphQL<{
    createProductSetupDraft: SetupDraft
  }>(
    `
      mutation CreateProductSetupDraft($input: CreateProductSetupDraftInput!) {
        createProductSetupDraft(input: $input) {
          id
          name
          partner
          baseCost
          retailPrice
          status
          notes
          createdAt
          updatedAt
        }
      }
    `,
    { input: payload }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.createProductSetupDraft }
}

export async function promoteProductSetupCandidate(
  payload: PromoteCandidatePayload
): Promise<ProductSetupResult<CatalogCandidate>> {
  const result = await postBackofficeGraphQL<{
    promoteProductSetupCandidate: CatalogCandidate
  }>(
    `
      mutation PromoteProductSetupCandidate(
        $input: PromoteProductSetupCandidateInput!
      ) {
        promoteProductSetupCandidate(input: $input) {
          id
          draftId
          title
          sku
          partner
          baseCost
          retailPrice
          estimatedMargin
          status
          channel
          updatedAt
          merchandisingNotes
          variants {
            id
            label
            color
            size
            status
          }
          artworkChecklist {
            frontArtwork
            backArtwork
            mockupReady
            printSpecChecked
          }
        }
      }
    `,
    { input: payload }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.promoteProductSetupCandidate }
}

export async function updateProductSetupCandidateStatus(
  id: string,
  status: string
): Promise<ProductSetupResult<CatalogCandidate>> {
  const result = await postBackofficeGraphQL<{
    updateProductSetupCandidateStatus: CatalogCandidate
  }>(
    `
      mutation UpdateProductSetupCandidateStatus($id: ID!, $status: String!) {
        updateProductSetupCandidateStatus(id: $id, status: $status) {
          id
          draftId
          title
          sku
          partner
          baseCost
          retailPrice
          estimatedMargin
          status
          channel
          updatedAt
          merchandisingNotes
          variants {
            id
            label
            color
            size
            status
          }
          artworkChecklist {
            frontArtwork
            backArtwork
            mockupReady
            printSpecChecked
          }
        }
      }
    `,
    { id, status }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateProductSetupCandidateStatus }
}
