package domain

import "path"

func evaluatePolicyStatements(
	request AccessRequest,
	statements []PolicyStatement,
) bool {
	action := request.Action
	resource := request.Resource
	if resource == "" {
		resource = "*"
	}

	allowed := false
	for _, statement := range statements {
		if !matchesPattern(statement.ActionPattern, action) {
			continue
		}
		if !matchesPattern(statement.ResourcePattern, resource) {
			continue
		}
		if statement.Effect == PolicyEffectDeny {
			return false
		}
		if statement.Effect == PolicyEffectAllow {
			allowed = true
		}
	}

	return allowed
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
