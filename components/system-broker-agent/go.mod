module github.com/kyma-project/kyma/components/system-broker-agent

go 1.14

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/Azure/go-autorest/autorest v0.9.6 // indirect
	github.com/kyma-incubator/compass v0.0.0-20200921061826-9f8cc0e16f01
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200818080816-8c81ea09adc7
	github.com/kyma-project/kyma/components/compass-runtime-agent v0.0.0-20200918054249-2504ff067c2b // indirect
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/client-go => k8s.io/client-go v0.18.8
)
