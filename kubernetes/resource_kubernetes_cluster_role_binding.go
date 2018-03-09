package kubernetes

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	rbacv1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func resourceKubernetesClusterRoleBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesClusterRoleBindingCreate,
		Read:   resourceKubernetesClusterRoleBindingRead,
		Exists: resourceKubernetesClusterRoleBindingExists,
		//Update: resourceKubernetesClusterRoleBindingUpdate,
		Delete: resourceKubernetesClusterRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("cluster role binding", true),
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
			"subject": {
				Type:        schema.TypeList,
				Description: "RoleRef contains information that points to the role being used",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_group": {
							Type:        schema.TypeString,
							Description: "APIGroup holds the API group of the referenced subject. Defaults to \"\" for ServiceAccount subjects. Defaults to \"rbac.authorization.k8s.io\" for User and Group subjects.",
							Optional:    true,
						},
						"kind": {
							Type:        schema.TypeString,
							Description: "Kind of object being referenced. Values defined by this API group are \"User\", \"Group\", and \"ServiceAccount\". If the Authorizer does not recognized the kind value, the Authorizer should report an error.",
							Required:    true,
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

func expandRoleRef(in map[string]interface{}) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: in["api_group"].(string),
		Kind:     in["kind"].(string),
		Name:     in["name"].(string),
	}
}

func expandSubjects(subjs []interface{}) []rbacv1.Subject {
	if len(subjs) == 0 {
		return []rbacv1.Subject{}
	}
	su := make([]rbacv1.Subject, len(subjs))

	for i, v := range subjs {
		sub := v.(map[string]interface{})

		su[i].APIGroup = sub["api_group"].(string)
		su[i].Kind = sub["kind"].(string)
		su[i].Name = sub["name"].(string)
		su[i].Namespace = sub["namespace"].(string)
	}

	return su
}

func resourceKubernetesClusterRoleBindingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	role_ref := expandRoleRef(d.Get("role_ref").(map[string]interface{}))
	subjects := expandSubjects(d.Get("subject").([]interface{}))

	clusterRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metadata,
		RoleRef:    role_ref,
		Subjects:   subjects,
	}
	log.Printf("[INFO] Creating new cluster role binding map: %#v", clusterRoleBinding)
	out, err := conn.RbacV1beta1().ClusterRoleBindings().Create(&clusterRoleBinding)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new cluster role binding: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesClusterRoleBindingRead(d, meta)
}

func resourceKubernetesClusterRoleBindingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading cluster role binding %s", name)
	crb, err := conn.RbacV1beta1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received cluster role binding: %#v", crb)
	err = d.Set("metadata", flattenMetadata(crb.ObjectMeta, d))
	if err != nil {
		return err
	}
	//d.Set("data", cfgMap.Data)

	return nil
}

func resourceKubernetesClusterRoleBindingExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetes.Clientset)

	_, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking cluster role binding %s", name)
	_, err = conn.RbacV1beta1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}

func resourceKubernetesClusterRoleBindingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	_, name, err := idParts(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting cluster role binding: %#v", name)
	err = conn.RbacV1beta1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("[INFO] Cluster role binding %s deleted", name)

	d.SetId("")
	return nil
}
