package entity

type AccessRequest struct {
	TenantID   string            `json:"tenant_id,omitempty"`
	OrgID      string            `json:"org_id,omitempty"`
	UserID     uint              `json:"user_id,omitempty"`
	Action     string            `json:"action"`
	Resource   string            `json:"resource"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type SimulateAccessInput struct {
	Scope            string            `json:"scope"`
	TenantID         string            `json:"tenant_id,omitempty"`
	UserID           uint              `json:"user_id"`
	Action           string            `json:"action"`
	Resource         string            `json:"resource"`
	UseAssumedRole   bool              `json:"use_assumed_role"`
	AssumedRole      *AssumedRole      `json:"assumed_role,omitempty"`
	SessionPolicy    []PolicyStatement `json:"session_policy,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
	ServicePrincipal string            `json:"service_principal,omitempty"`
}

type SimulateMatchedStatement struct {
	PolicyName      string            `json:"policy_name"`
	Effect          string            `json:"effect"`
	ActionPattern   string            `json:"action_pattern"`
	ResourcePattern string            `json:"resource_pattern"`
	Conditions      []PolicyCondition `json:"conditions,omitempty"`
	Source          string            `json:"source"`
}

type SimulateDecisionLayer struct {
	Layer             string                     `json:"layer"`
	Allowed           bool                       `json:"allowed"`
	Reason            string                     `json:"reason"`
	MatchedStatements []SimulateMatchedStatement `json:"matched_statements,omitempty"`
}

type SimulateAccessResult struct {
	Allowed           bool                       `json:"allowed"`
	DecisionSource    string                     `json:"decision_source"`
	Reason            string                     `json:"reason"`
	MatchedStatements []SimulateMatchedStatement `json:"matched_statements,omitempty"`
	Layers            []SimulateDecisionLayer    `json:"layers,omitempty"`
}
