package entity

import "fmt"

func BuildConsulKey(tenantID string, infraType InfraType, name string) string {
	if name == "" {
		name = "default"
	}
	return fmt.Sprintf("podzone/tenants/%s/connections/%s/%s", tenantID, infraType, name)
}
