export type CreateGroupFormValues = {
  scope: string;
  tenantId: string;
  name: string;
  description: string;
};

export type GroupMemberFormValues = {
  userId: string;
};

export type GroupPolicyAttachmentFormValues = {
  policyName: string;
};

export type GroupInlinePolicyFormValues = {
  name: string;
  description: string;
  statementsJson: string;
};
