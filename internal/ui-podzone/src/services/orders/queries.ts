import { postBackofficeGraphQL } from '../backofficeGraphql'
import { routedOrderFields } from './graphql'
import type {
  OrdersResult,
  RoutedOrder,
  RoutedOrderActivityFeedPage,
  RoutedOrderActivityFeedQuery,
  RoutedOrderRecommendation,
  RoutedOrderRecommendationQuery,
} from './types'

export async function getRoutedOrders(): Promise<
  OrdersResult<{ orders: RoutedOrder[] }>
> {
  const result = await postBackofficeGraphQL<{ routedOrders: RoutedOrder[] }>(`
    query RoutedOrders {
      routedOrders {
${routedOrderFields}
      }
    }
  `)
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: { orders: result.data.routedOrders || [] } }
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
