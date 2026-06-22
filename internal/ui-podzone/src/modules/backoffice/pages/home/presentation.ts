export function parseMoney(value: string) {
  const cleaned = value.replace(/[^0-9.]/g, '');
  const parsed = Number.parseFloat(cleaned);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

export function isOverdue(value?: string) {
  if (!value) {
    return false;
  }
  return new Date(value).getTime() < Date.now();
}

export function buildOrdersHref(
  tenantID: string,
  storeID: string,
  queueView: string,
  queueSort: string = 'priority',
  operatorLens?: string
) {
  const params = new URLSearchParams({
    queueView,
    queueSort,
  });
  if (operatorLens) {
    params.set('operatorLens', operatorLens);
  }
  if (storeID) {
    params.set('storeId', storeID);
  }
  return `/t/${tenantID}/orders?${params.toString()}`;
}

export function buildTenantHref(
  tenantID: string,
  storeID: string,
  path: string
) {
  const params = new URLSearchParams();
  if (storeID) {
    params.set('storeId', storeID);
  }
  const query = params.toString();
  return `/t/${tenantID}${path}${query ? `?${query}` : ''}`;
}
