package synchronization

import (
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
	//"github.com/pivotal-cf/brokerapi/v7/domain"
)

const (
	SpecAPIType    = "API"
	SpecEventsType = "Events"
)

func toClusterSystem(service osb.Service) v1alpha1.ClusterSystem {

	clusterServices := make([]v1alpha1.Service, 0)
	for _, plan := range service.Plans {
		clusterServices = append(clusterServices, toService(plan))
	}

	return v1alpha1.ClusterSystem{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterSystem",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: v1alpha1.ClusterSystemSpec{
			Services: clusterServices,
			CompassMetadata: &v1alpha1.CompassMetadata{
				ApplicationID: service.ID,
			},
		},
	}
}

func toService(plan osb.Plan) v1alpha1.Service {

	metadata := make([]map[string]interface{}, 0)

	val, ok := plan.Metadata["specifications"]
	if ok {
		if val != nil {
			array := val.([]interface{})

			for _, item := range array {
				metadata = append(metadata, item.(map[string]interface{}))
			}
		}
	}

	return v1alpha1.Service{
		ID:          plan.ID,
		Identifier:  "",        // Application Registry specific?
		Name:        plan.Name, //
		DisplayName: "",
		Description: plan.Description,
		Entries:     toEntries(metadata),

		// New fields used by V2 version
		//AuthCreateParameterSchema *string `json:"authCreateParameterSchema,omitempty"`
	}
}

func toEntries(metadata []map[string]interface{}) []v1alpha1.Entry {

	entries := make([]v1alpha1.Entry, 0)
	for _, item := range metadata {
		definitionID := item["definition_id"].(string)                   // APIDefinition.ID or EventDefinition.ID
		definitionName := item["definition_name"].(string)               // APIDefinition.Name or EventDefinition.Name
		specificationCategory := item["specification_category"].(string) // "api_definition" or "event_definition"
		specificationType := item["specification_type"].(string)         // APIDefinition.Spec.Type or EventAPIDefinition.Spec.Type
		// TODO: consider adding some fields
		//specificationFormat := item["specification_format"].(string) //APIDefinition.Spec.Format or //EventDefinition.Spec.Format
		specificationURL := item["specification_url"].(string) //APIDefinition.Spec.Format or //EventDefinition.Spec.Format

		// TODO
		//specificationUrl := item["specification_url"] //
		var entry v1alpha1.Entry

		switch specificationCategory {
		case "api_definition":
			entry = v1alpha1.Entry{
				ID:               definitionID,
				Name:             definitionName,
				Type:             SpecAPIType, //???
				TargetUrl:        "",          // TODO missing in Broker
				SpecificationUrl: specificationURL,
				ApiType:          specificationType,
			}
		case "event_definition":
			entry = v1alpha1.Entry{
				ID:               definitionID,
				Name:             definitionName,
				Type:             SpecEventsType, //???
				SpecificationUrl: specificationURL,
				ApiType:          specificationType,
			}
		default:
			logrus.Errorf("Unknown specification category %s", specificationCategory)
		}

		entries = append(entries, entry)
	}

	return entries
}
