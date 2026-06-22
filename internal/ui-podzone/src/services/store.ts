import { postBackofficeGraphQL } from './backofficeGraphql';

export type StoreInfo = {
  id: string;
  name: string;
  ownerId: string;
  isActive: boolean;
  description: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

type StoreResult<T> =
  | { success: true; data: T }
  | { success: false; message: string };

export async function listStores(): Promise<StoreResult<StoreInfo[]>> {
  const result = await postBackofficeGraphQL<{ stores: StoreInfo[] }>(
    `
    query Stores {
      stores {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    undefined,
    { includeStoreHeader: false }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.stores || [] };
}

export async function createStore(input: {
  name: string;
  description?: string;
}): Promise<StoreResult<StoreInfo>> {
  const result = await postBackofficeGraphQL<{ createStore: StoreInfo }>(
    `
    mutation CreateStore($input: CreateStoreInput!) {
      createStore(input: $input) {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    {
      input: {
        name: input.name,
        description: input.description || '',
      },
    },
    { includeStoreHeader: false }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.createStore };
}

export async function activateStore(
  id: string
): Promise<StoreResult<StoreInfo>> {
  const result = await postBackofficeGraphQL<{ activateStore: StoreInfo }>(
    `
    mutation ActivateStore($id: ID!) {
      activateStore(id: $id) {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    { id },
    { includeStoreHeader: false }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.activateStore };
}
