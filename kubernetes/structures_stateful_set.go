package kubernetes

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func flattenStatefulSetSpec(in appsv1.StatefulSetSpec, d *schema.ResourceData) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Replicas != nil {
		att["replicas"] = *in.Replicas
	}
	att["pod_management_policy"] = in.PodManagementPolicy
	if in.RevisionHistoryLimit != nil {
		att["revision_history_limit"] = *in.RevisionHistoryLimit
	}
	att["service_name"] = in.ServiceName
	att["selector"] = in.Selector.MatchLabels
	att["update_strategy"] = flattenStatefulSetUpdateStrategy(in.UpdateStrategy, d)

	templateMetadata := flattenMetadata(in.Template.ObjectMeta, d)
	podSpec, err := flattenPodSpec(in.Template.Spec)
	if err != nil {
		return nil, err
	}
	template := make(map[string]interface{})
	template["metadata"] = templateMetadata
	template["spec"] = podSpec
	att["template"] = []interface{}{template}

	volClaimTemplates := make([]map[string]interface{}, len(in.VolumeClaimTemplates), len(in.VolumeClaimTemplates))
	for i, claim := range in.VolumeClaimTemplates {
		claimState := make(map[string]interface{})
		claimState["metadata"] = flattenSubMetadata(claim.ObjectMeta, d, fmt.Sprintf("spec.0.volume_claim_templates.%d", i))
		claimState["spec"] = flattenPersistentVolumeClaimSpec(claim.Spec)
		claimState["use_default_provisioning"] = d.Get(fmt.Sprintf("spec.0.volume_claim_templates.%d.use_default_provisioning", i)).(bool)
		claimState["wait_until_bound"] = d.Get(fmt.Sprintf("spec.0.volume_claim_templates.%d.wait_until_bound", i)).(bool)
		volClaimTemplates[i] = claimState
	}
	att["volume_claim_templates"] = volClaimTemplates

	return []interface{}{att}, nil
}

func flattenStatefulSetUpdateStrategy(in appsv1.StatefulSetUpdateStrategy, d *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.RollingUpdate != nil {
		att["rolling_update"] = flattenStatefulSetStrategyRollingUpdate(in.RollingUpdate)
	}
	return []interface{}{att}
}

func flattenStatefulSetStrategyRollingUpdate(in *appsv1.RollingUpdateStatefulSetStrategy) []interface{} {
	att := make(map[string]interface{})
	if in.Partition != nil {
		att["partition"] = int(*in.Partition)
	}

	return []interface{}{att}
}

//
// EXPANDERS
//

func expandStatefulSetSpec(statefulSet []interface{}) (appsv1.StatefulSetSpec, error) {
	obj := appsv1.StatefulSetSpec{}
	if len(statefulSet) == 0 || statefulSet[0] == nil {
		return obj, nil
	}
	in := statefulSet[0].(map[string]interface{})

	if v, ok := in["update_strategy"]; ok {
		obj.UpdateStrategy = expandStatefulSetUpdateStrategy(v.([]interface{}))
	}

	obj.Replicas = ptrToInt32(int32(in["replicas"].(int)))
	obj.Selector = &metav1.LabelSelector{
		MatchLabels: expandStringMap(in["selector"].(map[string]interface{})),
	}
	obj.ServiceName = in["service_name"].(string)

	for _, v := range in["template"].([]interface{}) {
		template := v.(map[string]interface{})
		podSpec, err := expandPodSpec(template["spec"].([]interface{}))
		if err != nil {
			return obj, err
		}
		obj.Template = v1.PodTemplateSpec{
			Spec: podSpec,
		}

		if metaCfg, ok := template["metadata"]; ok {
			metadata := expandMetadata(metaCfg.([]interface{}))
			obj.Template.ObjectMeta = metadata
		}
	}

	volClaimTemplates := in["volume_claim_templates"].([]interface{})
	pvcTemplates := make([]v1.PersistentVolumeClaim, len(volClaimTemplates), len(volClaimTemplates))
	for i, claimTemplateRaw := range volClaimTemplates {
		claimTemplateConfig := claimTemplateRaw.(map[string]interface{})
		metadata := expandMetadata(claimTemplateConfig["metadata"].([]interface{}))
		use_default_provisioning := false
		if v, ok := claimTemplateConfig["use_default_provisioning"].(bool); ok {
			use_default_provisioning = v
		}
		pvcSpec, _ := expandPersistentVolumeClaimSpec(claimTemplateConfig["spec"].([]interface{}), use_default_provisioning)
		claim := v1.PersistentVolumeClaim{
			ObjectMeta: metadata,
			Spec:       pvcSpec,
		}
		pvcTemplates[i] = claim
	}
	obj.VolumeClaimTemplates = pvcTemplates

	return obj, nil
}

func expandStatefulSetUpdateStrategy(p []interface{}) appsv1.StatefulSetUpdateStrategy {
	obj := appsv1.StatefulSetUpdateStrategy{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"]; ok {
		obj.Type = appsv1.StatefulSetUpdateStrategyType(v.(string))
	}
	if obj.Type == appsv1.RollingUpdateStatefulSetStrategyType {
		if v, ok := in["rolling_update"]; ok {
			obj.RollingUpdate = expandRollingUpdateStatefulSetStrategy(v.([]interface{}))
		}
	}
	return obj
}

func expandRollingUpdateStatefulSetStrategy(p []interface{}) *appsv1.RollingUpdateStatefulSetStrategy {
	obj := appsv1.RollingUpdateStatefulSetStrategy{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["partition"]; ok {
		obj.Partition = ptrToInt32(int32(v.(int)))
	}

	return &obj
}
