package kubernetes

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

func expandRules(in []interface{}) []rbacv1.PolicyRule {
	if len(in) == 0 {
		return []rbacv1.PolicyRule{}
	}

	rules := make([]rbacv1.PolicyRule, len(in))

	for i, v := range in {
		rule := v.(map[string]interface{})

		if verbs, ok := rule["verbs"].([]interface{}); ok {
			rules[i].Verbs = expandStringSlice(verbs)
		}
		if api_groups, ok := rule["api_groups"].([]interface{}); ok {
			rules[i].APIGroups = expandStringSlice(api_groups)
		}
		if resources, ok := rule["resources"].([]interface{}); ok {
			rules[i].Resources = expandStringSlice(resources)
		}
		if resource_names, ok := rule["resource_names"].([]interface{}); ok {
			rules[i].ResourceNames = expandStringSlice(resource_names)
		}
		if non_resource_urls, ok := rule["non_resource_urls"].([]interface{}); ok {
			rules[i].NonResourceURLs = expandStringSlice(non_resource_urls)
		}
	}
	return rules
}

func flattenRules(in []rbacv1.PolicyRule) []interface{} {
	rules := make([]interface{}, len(in))
	for i, v := range in {
		rule := make(map[string]interface{})
		rule["verbs"] = v.Verbs
		rule["api_groups"] = v.APIGroups
		rule["resources"] = v.Resources
		rule["resource_names"] = v.ResourceNames
		rule["non_resource_urls"] = v.NonResourceURLs
		rules[i] = rule
	}
	return rules
}

func expandRoleRef(in map[string]interface{}) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: in["api_group"].(string),
		Kind:     in["kind"].(string),
		Name:     in["name"].(string),
	}
}

func flattenRoleRef(in rbacv1.RoleRef) interface{} {
	att := make(map[string]interface{})
	att["api_group"] = in.APIGroup
	att["kind"] = in.Kind
	att["name"] = in.Name

	return att
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
