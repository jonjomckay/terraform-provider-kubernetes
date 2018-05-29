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

func resourceKubernetesClusterRoleBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesClusterRoleBindingCreate,
		Read:   resourceKubernetesClusterRoleBindingRead,
		Exists: resourceKubernetesClusterRoleBindingExists,
		Update: resourceKubernetesClusterRoleBindingUpdate,
		Delete: resourceKubernetesClusterRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": rbacMetadataSchema("cluster role binding", true),
			"role_ref": {
				Type:        schema.TypeMap,
				Description: "RoleRef contains information that points to the role being used",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_group": {
							Type:        schema.TypeString,
							Description: "APIGroup holds the API group of the referenced subject. Defaults to \"\" for ServiceAccount subjects. Defaults to \"rbac.authorization.k8s.io\" for User and Group subjects.",
							Required:    true,
						},
						"kind": {
							Type:        schema.TypeString,
							Description: "Kind is the type of resource being referenced",
							Required:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name is the name of resource being referenced",
							Required:    true,
						},
					},
				},
			},
			"subject": {
				Type:        schema.TypeList,
				Description: "Subjects holds references to the objects the role applies to",
				Required:    true,
				ForceNew:    false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:        schema.TypeString,
							Description: "Kind of object being referenced. Values defined by this API group are \"User\", \"Group\", and \"ServiceAccount\". If the Authorizer does not recognized the kind value, the Authorizer should report an error.",
							Required:    true,
						},
						"api_group": {
							Type:        schema.TypeString,
							Description: "APIGroup holds the API group of the referenced subject. Defaults to \"\" for ServiceAccount subjects. Defaults to \"rbac.authorization.k8s.io\" for User and Group subjects.",
							Optional:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the object being referenced.",
							Required:    true,
						},
						"namespace": {
							Type:        schema.TypeString,
							Description: "Namespace of the referenced object. If the object kind is non-namespace, such as \"User\" or \"Group\", and this value is not empty the Authorizer should report an error.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesClusterRoleBindingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	clusterRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: expandMetadata(d.Get("metadata").([]interface{})),
		RoleRef:    expandRoleRef(d.Get("role_ref").(map[string]interface{})),
		Subjects:   expandSubjects(d.Get("subject").([]interface{})),
	}
	log.Printf("[INFO] Creating new cluster role binding map: %#v", clusterRoleBinding)
	out, err := conn.RbacV1().ClusterRoleBindings().Create(&clusterRoleBinding)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new cluster role binding: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesClusterRoleBindingRead(d, meta)
}

func resourceKubernetesClusterRoleBindingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading cluster role binding %s", name)
	crb, err := conn.RbacV1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received cluster role binding: %#v", crb)
	err = d.Set("metadata", flattenMetadata(crb.ObjectMeta, d))
	if err != nil {
		return err
	}

	err = d.Set("role_ref", flattenRoleRef(crb.RoleRef))
	if err != nil {
		return err
	}

	err = d.Set("subject", flattenSubjects(crb.Subjects))
	if err != nil {
		return err
	}

	return nil
}

func resourceKubernetesClusterRoleBindingExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking cluster role binding %s", name)
	_, err = conn.RbacV1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}

func resourceKubernetesClusterRoleBindingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)

	if d.HasChange("role_ref") {
		return fmt.Errorf("Failed to update cluster role binding: cannot change role ref")
	}

	if d.HasChange("subject") {
		subjects := expandSubjects(d.Get("subject").([]interface{}))
		ops = append(ops, &ReplaceOperation{
			Path:  "/subjects",
			Value: subjects,
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating cluster role binding %s", name)
	out, err := conn.RbacV1().ClusterRoleBindings().Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update cluster role binding: %s", err)
	}
	log.Printf("[INFO] Submitted updated cluster role binding %#v", out)

	return resourceKubernetesClusterRoleBindingRead(d, meta)
}

func resourceKubernetesClusterRoleBindingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetesProvider).conn

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting cluster role binding: %#v", name)
	err = conn.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("[INFO] Cluster role binding %s deleted", name)

	d.SetId("")
	return nil
}
