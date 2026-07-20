export type TenantMembership = {
  tenantId: string;
  userId: number;
  roleId?: number;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export function formatRoleLabel(roleName: string): string {
  return roleName
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}
