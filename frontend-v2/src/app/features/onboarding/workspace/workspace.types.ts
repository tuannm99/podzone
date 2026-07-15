export type TenantMembership = {
  tenantId: string;
  userId: number;
  roleId?: number;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};
