package kubernetes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"k8s.io/api/apps/v1"
)

func TestAccKubernetesStatefulSet_basic(t *testing.T) {
	var sset v1.StatefulSet

	statefulSetName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	imageName1 := "nginx:1.7.9"
	imageName2 := "nginx:1.11"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesStatefulSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesStatefulSetConfig_basic(statefulSetName, imageName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.labels.%", "1"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.labels.app", "one"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.service_name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.template.0.spec.0.container.0.image", imageName1),
				),
			},
			{
				Config: testAccKubernetesStatefulSetConfig_basic(statefulSetName, imageName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.template.0.spec.0.container.0.image", imageName2),
				),
			},
		},
	})
}

func TestAccKubernetesStatefulSet_pvcTemplate(t *testing.T) {
	var sset v1.StatefulSet

	statefulSetName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	imageName1 := "nginx:1.7.9"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesStatefulSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesStatefulSetConfig_pvcTemplate(statefulSetName, imageName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.labels.%", "0"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.service_name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.template.0.spec.0.container.0.image", imageName1),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.template.0.spec.0.container.0.image", imageName1),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.volume_claim_templates.#", "1"),
				),
			},
		},
	})
}

func TestAccKubernetesStatefulSet_updateStrategy(t *testing.T) {
	var sset v1.StatefulSet

	statefulSetName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesStatefulSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesStatefulSetConfig_updateStrategy(statefulSetName, "RollingUpdate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.labels.%", "0"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_stateful_set.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.service_name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.0.type", "RollingUpdate"),
				),
			},
			{
				Config: testAccKubernetesStatefulSetConfig_updateStrategy(statefulSetName, "OnDelete"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.service_name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.0.type", "OnDelete"),
				),
			},
			{
				Config: testAccKubernetesStatefulSetConfig_updateStrategyRollingUpdate(statefulSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesStatefulSetExists("kubernetes_stateful_set.test", &sset),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "metadata.0.name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.service_name", statefulSetName),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.0.type", "RollingUpdate"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.0.rolling_update.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_stateful_set.test", "spec.0.update_strategy.0.rolling_update.0.partition", "1"),
				),
			},
		},
	})
}

func testAccCheckKubernetesStatefulSetExists(n string, obj *v1.StatefulSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		kp := testAccProvider.Meta().(*kubernetesProvider)

		namespace, name, _ := idParts(rs.Primary.ID)
		out, err := readStatefulSet(kp, namespace, name)
		if err != nil {
			return err
		}
		*obj = *out
		return nil
	}
}

func testAccCheckKubernetesStatefulSetDestroy(s *terraform.State) error {
	kp := testAccProvider.Meta().(*kubernetesProvider)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_stateful_set" {
			continue
		}
		namespace, name, _ := idParts(rs.Primary.ID)
		resp, err := readStatefulSet(kp, namespace, name)
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("Stateful Set still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccKubernetesStatefulSetConfig_basic(name, image string) string {
	return fmt.Sprintf(`
resource "kubernetes_stateful_set" "test" {
  metadata {
		name = "%s"
		labels {
			app = "one"
		}
  }
  spec {
    replicas = 2
    selector {
      app = "one"
    }
    service_name = "%s"
    template {
			metadata {
				labels {
					app = "one"
				}
			}
			spec {
				container {
					image = "%s"
					name  = "tf-acc-test"
				}
			}
    }
  }
}
`, name, name, image)
}

func testAccKubernetesStatefulSetConfig_pvcTemplate(name, image string) string {
	return fmt.Sprintf(`
resource "kubernetes_stateful_set" "test" {
  metadata {
		name = "%s"
  }
  spec {
    replicas = 2
    selector {
      app = "one"
    }
    service_name = "%s"
    template {
			metadata {
				labels {
					app = "one"
				}
			}
			spec {
				container {
					image = "%s"
					name  = "tf-acc-test"
				}
			}
    }

		volume_claim_templates {
			metadata {
				name = "pvc"
			}
			spec {
				access_modes = ["ReadWriteOnce"]
				resources {
					requests {
						storage = "2Gi"
					}
				}
			}
		}
  }
}
`, name, name, image)
}

func testAccKubernetesStatefulSetConfig_updateStrategy(name, strategy string) string {
	return fmt.Sprintf(`
resource kubernetes_stateful_set "test" {
  metadata {
    name = "%s"
  }

  spec {
    selector {
      app = "pinger"
    }

    service_name = "%s"

    update_strategy {
      type = "%s"
    }

    replicas = 2

    template {
      metadata {
        labels {
          app = "pinger"
        }
      }

      spec {
        termination_grace_period_seconds = 5

        container {
          name = "pinger-a"
          image = "debian:buster"
          command = ["ping", "github.com"]
        }
      }
    }
  }
}
`, name, name, strategy)
}

func testAccKubernetesStatefulSetConfig_updateStrategyRollingUpdate(name string) string {
	return fmt.Sprintf(`
resource kubernetes_stateful_set "test" {
  metadata {
    name = "%s"
  }

  spec {
    selector {
      app = "pinger"
    }

    service_name = "%s"

    update_strategy {
      type = "RollingUpdate"
	  rolling_update {
        partition = 1
      }
    }

    replicas = 2

    template {
      metadata {
        labels {
          app = "pinger"
        }
      }

      spec {
        termination_grace_period_seconds = 5

        container {
          name = "pinger-a"
          image = "debian:buster"
          command = ["ping", "github.com"]
        }
      }
    }
  }
}
`, name, name)
}
