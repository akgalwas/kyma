package synchronization

import (
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

func toClusterSystem(service osb.Service) v1alpha1.ClusterSystem {
	return v1alpha1.ClusterSystem{}
}

func toService(plan osb.Plan) v1alpha1.Service {
	return v1alpha1.Service{
		ID:          plan.ID,
		Identifier:  "",        // Application Registry specific?
		Name:        plan.Name, //
		DisplayName: "",
		Description: plan.Description,
		//Entries:     []Entry `json:"entries"`

		// New fields used by V2 version
		//AuthCreateParameterSchema *string `json:"authCreateParameterSchema,omitempty"`
	}
}

//func toEntries(metadata []map[string]interface{}) []v1alpha1.Entry {
//
//	for _, item := range metadata {
//		definitionID := item["definition_id"] // APIDefinition.ID or EventDefinition.ID
//		definitionName := item["definition_name"] // APIDefinition.Name or EventDefinition.Name
//		specificationCategory := item["specification_category"] // "api_definition" or "event_definition"
//		specificationType := item["specification_type"] // APIDefinition.Spec.Type or EventAPIDefinition.Spec.Type
//		specificationFormat := item["specification_format"] //
//	}
//
//	specifications["definition_id"] = apiDef.ID
//	specifications["definition_name"] = apiDef.Name
//	specifications["specification_category"] = "api_definition"
//	specifications["specification_type"] = apiDef.Spec.Type
//	specifications["specification_format"] = specsFormatHeader
//	specifications["specification_url"] =
//}
