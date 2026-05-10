import { TENANT_GQL_URL } from './baseurl';
import { http, type HttpError } from './http';

type GraphQLErrorItem = {
  message?: string;
};

type GraphQLPayload<T> = {
  data?: T;
  errors?: GraphQLErrorItem[];
};

function toMessage(error: unknown, fallback: string): string {
  if (
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
  ) {
    return error.message;
  }
  return fallback;
}

export async function postBackofficeGraphQL<T>(
  query: string,
  variables?: Record<string, unknown>
): Promise<{ success: true; data: T } | { success: false; message: string }> {
  try {
    const { data } = await http.post<GraphQLPayload<T>>(TENANT_GQL_URL, {
      query,
      variables,
    });

    if (data.errors && data.errors.length > 0) {
      return {
        success: false,
        message: data.errors[0]?.message || 'GraphQL request failed',
      };
    }

    if (!data.data) {
      return {
        success: false,
        message: 'GraphQL request returned no data',
      };
    }

    return { success: true, data: data.data };
  } catch (error) {
    return {
      success: false,
      message: toMessage(error as HttpError, 'GraphQL request failed'),
    };
  }
}
