package kubernetes

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	rbac_v1 "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccKubernetesClusterRole_basic(t *testing.T) {
	var role rbac_v1.ClusterRole
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_cluster_role.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesClusterRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRoleConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesClusterRoleExists("kubernetes_cluster_role.test", &role),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.TestAnnotationTwo", "two"),
					testAccCheckMetaAnnotations(&role.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.%", "3"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.TestLabelTwo", "two"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&role.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelTwo": "two", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.0", "get"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.1", "list"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.api_groups.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.api_groups.0", ""),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.resources.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.resources.0", "configmaps"),
				),
			},
			{
				Config: testAccKubernetesClusterRoleConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesClusterRoleExists("kubernetes_cluster_role.test", &role),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.Different", "1234"),
					testAccCheckMetaAnnotations(&role.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "Different": "1234"}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.%", "2"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&role.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.#", "3"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.0", "get"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.1", "list"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.verbs.2", "write"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.api_groups.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.api_groups.0", ""),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.resources.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.resources.0", "configmaps"),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "rule.0.resources.1", "secrets"),
				),
			},
		},
	})
}

func TestAccKubernetesClusterRole_importBasic(t *testing.T) {
	resourceName := "kubernetes_cluster_role.test"
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRoleConfig_basic(name),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccKubernetesClusterRole_generatedName(t *testing.T) {
	var role rbac_v1.ClusterRole
	prefix := "tf-acc-test-gen-"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_cluster_role.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesClusterRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRoleConfig_generatedName(prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesClusterRoleExists("kubernetes_cluster_role.test", &role),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.annotations.%", "0"),
					testAccCheckMetaAnnotations(&role.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.labels.%", "0"),
					testAccCheckMetaLabels(&role.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("kubernetes_cluster_role.test", "metadata.0.generate_name", prefix),
					resource.TestMatchResourceAttr("kubernetes_cluster_role.test", "metadata.0.name", regexp.MustCompile("^"+prefix)),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_cluster_role.test", "metadata.0.uid"),
				),
			},
		},
	})
}

func TestAccKubernetesClusterRole_importGeneratedName(t *testing.T) {
	resourceName := "kubernetes_cluster_role.test"
	prefix := "tf-acc-test-gen-import-"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRoleConfig_generatedName(prefix),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKubernetesClusterRoleDestroy(s *terraform.State) error {
	kp := testAccProvider.Meta().(*kubernetesProvider)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_cluster_role" {
			continue
		}
		_, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		resp, err := kp.conn.RbacV1().ClusterRoles().Get(name, meta_v1.GetOptions{})
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("ClusterRole still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckKubernetesClusterRoleExists(n string, obj *rbac_v1.ClusterRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		kp := testAccProvider.Meta().(*kubernetesProvider)
		_, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		out, err := kp.conn.RbacV1().ClusterRoles().Get(name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccKubernetesClusterRoleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cluster_role" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			TestAnnotationTwo = "two"
		}
		labels {
			TestLabelOne = "one"
			TestLabelTwo = "two"
			TestLabelThree = "three"
		}
		name = "%s"
	}
	rule {
	    api_groups = [""]
	    resources = ["configmaps"]
	    verbs = ["get", "list"]
	}
}`, name)
}

func testAccKubernetesClusterRoleConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cluster_role" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			Different = "1234"
		}
		labels {
			TestLabelOne = "one"
			TestLabelThree = "three"
		}
		name = "%s"
	}
	rule {
	    api_groups = [""]
	    resources = ["configmaps", "secrets"]
	    verbs = ["get", "list", "write"]
	}
}`, name)
}

func testAccKubernetesClusterRoleConfig_generatedName(prefix string) string {
	return fmt.Sprintf(`
resource "kubernetes_cluster_role" "test" {
	metadata {
		generate_name = "%s"
	}
	rule {
	    api_groups = [""]
	    resources = ["configmaps"]
	    verbs = ["get", "list"]
	}
}`, prefix)
}
