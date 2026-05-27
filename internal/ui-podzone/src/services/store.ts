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
  const result = await postBackofficeGraphQL<{ stores: StoreInfo[] }>(`
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
  `, undefined, { includeStoreHeader: false });
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.stores || [] };
}
