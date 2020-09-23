package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSystemSpec   `json:"spec"`
	Status            ClusterSystemStatus `json:"status,omitempty"`
}

func (app ClusterSystem) ShouldSkipInstallation() bool {
	return app.Spec.SkipInstallation == true
}

func (app ClusterSystem) GetClusterSystemID() string {
	if app.Spec.CompassMetadata == nil {
		return ""
	}

	return app.Spec.CompassMetadata.ApplicationID
}

type ClusterSystemStatus struct {
	// Represents the status of ClusterSystem release installation
	InstallationStatus InstallationStatus `json:"installationStatus"`
}

type InstallationStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (pw *ClusterSystem) GetObjectKind() schema.ObjectKind {
	return &ClusterSystem{}
}

// ApplicationSpec defines spec section of the Application custom resource
type ClusterSystemSpec struct {
	Description      string            `json:"description"`
	SkipInstallation bool              `json:"skipInstallation,omitempty"`
	Services         []Service         `json:"services"`
	Labels           map[string]string `json:"labels"`
	// TODO: Do we really need Tenant and group? It seems to be related to C4 Hana Cockpit
	Tenant          string           `json:"tenant,omitempty"`
	Group           string           `json:"group,omitempty"`
	CompassMetadata *CompassMetadata `json:"compassMetadata,omitempty"`

	// New fields used by V2 version TODO - remove this comment

	Tags                []string `json:"tags,omitempty"`
	DisplayName         string   `json:"displayName"`
	ProviderDisplayName string   `json:"providerDisplayName"`
	LongDescription     string   `json:"longDescription"`
}

type CompassMetadata struct {
	ApplicationID  string         `json:"applicationId"`
	Authentication Authentication `json:"authentication"`
}

type Authentication struct {
	ClientIds []string `json:"clientIds"`
}

// Entry defines, what is enabled by activating the service.
type Entry struct {
	Type                        string      `json:"type"`
	TargetUrl                   string      `json:"targetUrl"`
	SpecificationUrl            string      `json:"specificationUrl,omitempty"`
	ApiType                     string      `json:"apiType,omitempty"`
	Credentials                 Credentials `json:"credentials,omitempty"`
	RequestParametersSecretName string      `json:"requestParametersSecretName,omitempty"`

	// New fields used by V2 version TODO - remove this comment
	Name string `json:"name"`
	ID   string `json:"id"`
}

type CSRFInfo struct {
	TokenEndpointURL string `json:"tokenEndpointURL"`
}

// Credentials defines type of authentication and where the credentials are stored
type Credentials struct {
	Type              string    `json:"type"`
	SecretName        string    `json:"secretName"`
	AuthenticationUrl string    `json:"authenticationUrl,omitempty"`
	CSRFInfo          *CSRFInfo `json:"csrfInfo,omitempty"`
}

// Service represents part of the remote environment, which is mapped 1 to 1 in the service-catalog to:
// - service class in V1
// - service plans in V2 (since api-packages support)
type Service struct {
	ID          string  `json:"id"`
	Identifier  string  `json:"identifier"`
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Description string  `json:"description"`
	Entries     []Entry `json:"entries"`

	// New fields used by V2 version
	AuthCreateParameterSchema *string `json:"authCreateParameterSchema,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterSystem `json:"items"`
}
