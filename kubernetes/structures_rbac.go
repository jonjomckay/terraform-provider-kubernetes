package kubernetes

import (
	rbacv1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func flattenRoleRef(in rbacv1.RoleRef) interface{} {
	att := make(map[string]interface{})
	att["api_group"] = in.APIGroup
	att["kind"] = in.Kind
	att["name"] = in.Name

	return att
}

func flattenSubjects(in []rbacv1.Subject) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		obj := make(map[string]interface{})
		obj["kind"] = v.Kind
		if v.APIGroup != "" {
			obj["api_group"] = v.APIGroup
		}
		obj["name"] = v.Name
		obj["namespace"] = v.Namespace
		att[i] = obj
	}

	return att
}

func expandRoleRef(in map[string]interface{}) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: in["api_group"].(string),
		Kind:     in["kind"].(string),
		Name:     in["name"].(string),
	}
}

func expandSubjects(s []interface{}) []rbacv1.Subject {
	if len(s) == 0 {
		return []rbacv1.Subject{}
	}

	subjects := make([]rbacv1.Subject, len(s))

	for i, v := range s {
		subject := v.(map[string]interface{})

		if kind, ok := subject["kind"].(string); ok {
			subjects[i].Kind = kind
		}

		if api_group, ok := subject["api_group"].(string); ok {
			subjects[i].APIGroup = api_group
		}

		if name, ok := subject["name"].(string); ok {
			subjects[i].Name = name
		}

		if namespace, ok := subject["namespace"].(string); ok {
			subjects[i].Namespace = namespace
		}
	}

	return subjects
}
