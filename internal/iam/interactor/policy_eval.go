package interactor

import (
	"net"
	"path"
	"strconv"
	"strings"
	"time"
)

func evaluatePolicyStatements(
	request AccessRequest,
	statements []PolicyStatement,
) bool {
	result := explainPolicyStatements("", request, statements)
	return result.Allowed
}

func explainPolicyStatements(
	source string,
	request AccessRequest,
	statements []PolicyStatement,
) SimulateAccessResult {
	action := request.Action
	resource := request.Resource
	if resource == "" {
		resource = "*"
	}

	allowed := false
	matched := make([]SimulateMatchedStatement, 0)
	for _, statement := range statements {
		if !matchesPattern(statement.ActionPattern, action) {
			continue
		}
		if !matchesPattern(statement.ResourcePattern, resource) {
			continue
		}
		if !matchesConditions(statement.Conditions, request) {
			continue
		}
		matched = append(matched, SimulateMatchedStatement{
			PolicyName:      statement.PolicyName,
			Effect:          statement.Effect,
			ActionPattern:   statement.ActionPattern,
			ResourcePattern: statement.ResourcePattern,
			Conditions:      statement.Conditions,
			Source:          source,
		})
		if statement.Effect == PolicyEffectDeny {
			return SimulateAccessResult{
				Allowed:           false,
				DecisionSource:    source,
				Reason:            "explicit deny matched",
				MatchedStatements: matched,
			}
		}
		if statement.Effect == PolicyEffectAllow {
			allowed = true
		}
	}

	if allowed {
		return SimulateAccessResult{
			Allowed:           true,
			DecisionSource:    source,
			Reason:            "allow matched",
			MatchedStatements: matched,
		}
	}
	return SimulateAccessResult{
		Allowed:           false,
		DecisionSource:    source,
		Reason:            "no matching statement",
		MatchedStatements: matched,
	}
}

func matchesConditions(conditions []PolicyCondition, request AccessRequest) bool {
	if len(conditions) == 0 {
		return true
	}
	for _, condition := range conditions {
		if !matchesCondition(condition, request) {
			return false
		}
	}
	return true
}

func matchesCondition(condition PolicyCondition, request AccessRequest) bool {
	actual := requestAttribute(request, condition.Key)
	switch condition.Operator {
	case "", ConditionStringEquals:
		return actual == condition.Value
	case ConditionStringLike:
		return matchesPattern(condition.Value, actual)
	case ConditionStringNotEquals:
		return actual != condition.Value
	case ConditionStringNotLike:
		return !matchesPattern(condition.Value, actual)
	case ConditionBool:
		return strings.EqualFold(actual, condition.Value)
	case ConditionNumericEquals:
		return compareNumeric(actual, condition.Value) == 0
	case ConditionNumericGreaterThanEquals:
		return compareNumeric(actual, condition.Value) >= 0
	case ConditionNumericLessThanEquals:
		return compareNumeric(actual, condition.Value) <= 0
	case ConditionDateGreaterThan:
		return compareTime(actual, condition.Value) > 0
	case ConditionDateLessThan:
		return compareTime(actual, condition.Value) < 0
	case ConditionIpAddress:
		return matchesIPNet(actual, condition.Value)
	case ConditionNull:
		wantNull := strings.EqualFold(condition.Value, "true")
		return (actual == "") == wantNull
	default:
		return false
	}
}

func requestAttribute(request AccessRequest, key string) string {
	if request.Attributes != nil {
		if value, ok := request.Attributes[key]; ok {
			return value
		}
	}
	switch key {
	case "tenant_id":
		return request.TenantID
	case "org_id":
		return request.OrgID
	case "user_id":
		if request.UserID == 0 {
			return ""
		}
		return strconv.Itoa(int(request.UserID))
	case "action":
		return request.Action
	case "resource":
		return request.Resource
	default:
		if strings.HasPrefix(key, "aws:PrincipalTag/") || strings.HasPrefix(key, "principal_tag:") {
			tagKey := strings.TrimPrefix(strings.TrimPrefix(key, "aws:PrincipalTag/"), "principal_tag:")
			if request.Attributes != nil {
				return request.Attributes["principal_tag:"+tagKey]
			}
		}
		if strings.HasPrefix(key, "aws:RequestTag/") || strings.HasPrefix(key, "request_tag:") {
			tagKey := strings.TrimPrefix(strings.TrimPrefix(key, "aws:RequestTag/"), "request_tag:")
			if request.Attributes != nil {
				return request.Attributes["request_tag:"+tagKey]
			}
		}
		return ""
	}
}

func matchesPattern(pattern string, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	ok, err := path.Match(pattern, value)
	if err != nil {
		return pattern == value
	}
	return ok
}

func compareNumeric(actual string, expected string) int {
	a, errA := strconv.ParseFloat(actual, 64)
	b, errB := strconv.ParseFloat(expected, 64)
	if errA != nil || errB != nil {
		return -2
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareTime(actual string, expected string) int {
	a, errA := time.Parse(time.RFC3339, actual)
	b, errB := time.Parse(time.RFC3339, expected)
	if errA != nil || errB != nil {
		return -2
	}
	switch {
	case a.Before(b):
		return -1
	case a.After(b):
		return 1
	default:
		return 0
	}
}

func matchesIPNet(actual string, cidr string) bool {
	ip := net.ParseIP(strings.TrimSpace(actual))
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return false
	}
	return network.Contains(ip)
}
