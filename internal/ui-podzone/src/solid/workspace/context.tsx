import {
  createContext,
  createEffect,
  createMemo,
  createSignal,
  type ParentProps,
  useContext,
} from 'solid-js';
import { tenantStorage } from '../../services/tenantStorage';
import { listStores, type StoreInfo } from '../../services/store';
import { storeStorage } from '../../services/storeStorage';

type WorkspaceContextValue = {
  tenantId: () => string;
  stores: () => StoreInfo[];
  currentStoreId: () => string;
  currentStore: () => StoreInfo | undefined;
  loading: () => boolean;
  error: () => string;
  setCurrentStoreId: (storeId: string) => void;
};

const WorkspaceContext = createContext<WorkspaceContextValue>();

function pickInitialStore(tenantId: string, stores: StoreInfo[]) {
  const requested =
    new URLSearchParams(window.location.search).get('storeId') || '';
  const requestedMatch = stores.find((store) => store.id === requested.trim());
  if (requestedMatch) return requestedMatch.id;

  const persisted = storeStorage.getStoreID(tenantId);
  const persistedMatch = stores.find((store) => store.id === persisted);
  if (persistedMatch) return persistedMatch.id;
  return '';
}

function syncStoreIdToURL(storeId: string) {
  const normalizedStoreId = storeId.trim();
  if (!normalizedStoreId || !window.location.pathname.startsWith('/t/')) {
    return;
  }
  const params = new URLSearchParams(window.location.search);
  params.set('storeId', normalizedStoreId);
  const query = params.toString();
  const nextURL = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash}`;
  window.history.replaceState(window.history.state, '', nextURL);
}

export function TenantWorkspaceProvider(
  props: ParentProps<{ tenantId: string }>
) {
  const [stores, setStores] = createSignal<StoreInfo[]>([]);
  const [currentStoreId, setCurrentStoreIdState] = createSignal('');
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal('');

  const tenantId = createMemo(() => props.tenantId.trim());

  const setCurrentStoreId = (storeId: string) => {
    const normalizedStoreId = storeId.trim();
    if (!tenantId() || !normalizedStoreId) return;
    setCurrentStoreIdState(normalizedStoreId);
    storeStorage.setStoreID(tenantId(), normalizedStoreId);
    syncStoreIdToURL(normalizedStoreId);
  };

  createEffect(() => {
    const nextTenantId = tenantId();
    if (!nextTenantId) {
      setStores([]);
      setCurrentStoreIdState('');
      setError('');
      return;
    }

    tenantStorage.setTenantID(nextTenantId);
    setLoading(true);
    setError('');
    void listStores()
      .then((result) => {
        if (!result.success) {
          setStores([]);
          setCurrentStoreIdState('');
          setError(result.message);
          return;
        }
        setStores(result.data);
        const initialStoreId = pickInitialStore(nextTenantId, result.data);
        setCurrentStoreIdState(initialStoreId);
        if (initialStoreId) {
          storeStorage.setStoreID(nextTenantId, initialStoreId);
          syncStoreIdToURL(initialStoreId);
        }
      })
      .finally(() => {
        setLoading(false);
      });
  });

  const currentStore = createMemo(() =>
    stores().find((store) => store.id === currentStoreId())
  );

  return (
    <WorkspaceContext.Provider
      value={{
        tenantId,
        stores,
        currentStoreId,
        currentStore,
        loading,
        error,
        setCurrentStoreId,
      }}
    >
      {props.children}
    </WorkspaceContext.Provider>
  );
}

export function useTenantWorkspace() {
  return useContext(WorkspaceContext) || null;
}
