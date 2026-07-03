import { postBackofficeGraphQL } from '../backofficeGraphql'
import {
  normalizePageInfo,
  toGraphQLCollectionInput,
  type CollectionPage,
  type CollectionQuery,
  type WirePageInfo,
} from '../collection'
import { routedOrderFields } from './graphql'
import type {
  OrdersResult,
  RoutedOrder,
  RoutedOrderActivityFeedPage,
  RoutedOrderActivityFeedQuery,
  RoutedOrderRecommendation,
  RoutedOrderRecommendationQuery,
} from './types'

export async function getRoutedOrderPage(
  query: CollectionQuery
): Promise<OrdersResult<CollectionPage<RoutedOrder>>> {
  const result = await postBackofficeGraphQL<{
    routedOrders: {
      items: RoutedOrder[]
      pageInfo: WirePageInfo
    }
  }>(
    `
    query RoutedOrders($collection: CollectionInput) {
      routedOrders(collection: $collection) {
        items {
${routedOrderFields}
        }
        pageInfo {
          total
          page
          pageSize
          totalPages
          hasNext
          hasPrevious
        }
      }
    }
  `,
    {
      collection: toGraphQLCollectionInput(query),
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return {
    success: true,
    data: {
      items: result.data.routedOrders?.items || [],
      pageInfo: normalizePageInfo(result.data.routedOrders?.pageInfo, query),
    },
  }
}

export async function getRoutedOrders(): Promise<
  OrdersResult<{ orders: RoutedOrder[] }>
> {
  const orders: RoutedOrder[] = []
  for (let page = 1; ; page += 1) {
    const result = await getRoutedOrderPage({
      page,
      pageSize: 100,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    })
    if (!result.success) return result
    orders.push(...result.data.items)
    if (!result.data.pageInfo.hasNext) break
  }
  return { success: true, data: { orders } }
}

export async function getRoutedOrderActivities(
  input: RoutedOrderActivityFeedQuery
): Promise<OrdersResult<RoutedOrderActivityFeedPage>> {
  const result = await postBackofficeGraphQL<{
    routedOrderActivities: RoutedOrderActivityFeedPage
  }>(
    `
      query RoutedOrderActivities($input: RoutedOrderActivityFeedInput) {
        routedOrderActivities(input: $input) {
          entries {
            orderId
            productTitle
            partner
            operatorAssignee
            activity {
              type
              actor
              message
              details {
                key
                value
              }
              createdAt
            }
          }
          total
          nextCursor
        }
      }
    `,
    { input }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.routedOrderActivities }
}

export async function getRoutedOrderRecommendation(
  input: RoutedOrderRecommendationQuery
): Promise<OrdersResult<RoutedOrderRecommendation>> {
  const result = await postBackofficeGraphQL<{
    routedOrderRecommendation: RoutedOrderRecommendation
  }>(
    `
      query RoutedOrderRecommendation($input: RoutedOrderRecommendationInput!) {
        routedOrderRecommendation(input: $input) {
          candidateId
          productTitle
          candidatePartner
          productType
          shipRegion
          selectedPartner
          blockedReasonCode
          blockedReason
          summary
          options {
            eligible
            reason
            estimatedFulfillmentCost
            estimatedShippingCost
            estimatedUnitMargin
            partner {
              id
              code
              name
              partnerType
              status
              supportedProductTypes
              supportedRegions
              slaDays
              routingPriority
              baseFulfillmentCost
              shippingCostRules {
                region
                cost
              }
            }
          }
        }
      }
    `,
    { input }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.routedOrderRecommendation }
}
