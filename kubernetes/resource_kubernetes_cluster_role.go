package kubernetes

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
)

func resourceKubernetesClusterRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesClusterRoleCreate,
		Read:   resourceKubernetesClusterRoleRead,
		Exists: resourceKubernetesClusterRoleExists,
		Update: resourceKubernetesClusterRoleUpdate,
		Delete: resourceKubernetesClusterRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": rbacMetadataSchema("cluster role", true),
			"rule": {
				Type:        schema.TypeList,
				Description: "Policy Rules",
				Required:    true,
				ForceNew:    false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"verbs": {
							Type:        schema.TypeList,
							Description: "Verbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.  VerbAll represents all kinds.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Required:    true,
						},
						"api_groups": {
							Type:        schema.TypeList,
							Description: "APIGroups is the name of the APIGroup that contains the resources.  If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"resources": {
							Type:        schema.TypeList,
							Description: "Resources is a list of resources this rule applies to.  ResourceAll represents all resources.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"resource_names": {
							Type:        schema.TypeList,
							Description: "ResourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"non_resource_urls": {
							Type:        schema.TypeList,
							Description: "NonResourceURLs is a set of partial urls that a user should have access to.  *s are allowed, but only as the full, final step in the path. Since non-resource URLs are not namespaced, this field is only applicable for ClusterRoles referenced from a ClusterRoleBinding. Rules can either apply to API resources (such as \"pods\" or \"secrets\") or non-resource URL paths (such as \"/api\"),  but not both.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesClusterRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	clusterRole := rbacv1.ClusterRole{
		ObjectMeta: expandMetadata(d.Get("metadata").([]interface{})),
		Rules:      expandRules(d.Get("rule").([]interface{})),
	}
	log.Printf("[INFO] Creating new cluster role map: %#v", clusterRole)
	out, err := conn.RbacV1().ClusterRoles().Create(&clusterRole)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new cluster role: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesClusterRoleRead(d, meta)
}

func resourceKubernetesClusterRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading cluster role %s", name)
	crb, err := conn.RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received cluster role: %#v", crb)
	err = d.Set("metadata", flattenMetadata(crb.ObjectMeta, d))
	if err != nil {
		return err
	}

	err = d.Set("rule", flattenRules(crb.Rules))
	if err != nil {
		return err
	}

	return nil
}

func resourceKubernetesClusterRoleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking cluster role %s", name)
	_, err = conn.RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}

func resourceKubernetesClusterRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)

	//if d.HasChange("rule") {
	//	return fmt.Errorf("Failed to update cluster role: cannot change role ref")
	//}

	if d.HasChange("rule") {
		rules := expandRules(d.Get("rule").([]interface{}))
		ops = append(ops, &ReplaceOperation{
			Path:  "/rules",
			Value: rules,
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating cluster role %s", name)
	out, err := conn.RbacV1().ClusterRoles().Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update cluster role: %s", err)
	}
	log.Printf("[INFO] Submitted updated cluster role %#v", out)

	return resourceKubernetesClusterRoleRead(d, meta)
}

func resourceKubernetesClusterRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting cluster role: %#v", name)
	err = conn.RbacV1().ClusterRoles().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("[INFO] Cluster role %s deleted", name)

	d.SetId("")
	return nil
}
