package transformers

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"

	"github.com/davecgh/go-spew/spew"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
)

type annotationsTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestAnnotationsRun(t *testing.T) {
	var ingress = gvk.Gvk{Kind: "Ingress"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	for _, test := range []struct {
		name     string
		keys     []string
		input    *annotationsTransformerArgs
		expected *annotationsTransformerArgs
	}{
		{
			name: "it should remove matching annotations",
			keys: []string{
				"helm.sh/hook",
				"helm.sh/hook-weight",
				"remove-me",
			},
			input: &annotationsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(ingress, "ing1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Ingress",
								"metadata": map[string]interface{}{
									"name": "ing1",
									"annotations": map[string]interface{}{
										"kubernetes.io/ingress.class": "nginx",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"annotations": map[string]interface{}{
										"helm.sh/hook":        "pre-install",
										"helm.sh/hook-weight": "5",
									},
								},
								"spec": map[string]interface{}{
									"template": map[string]interface{}{
										"metadata": map[string]interface{}{
											"annotations": map[string]interface{}{
												"iam.amazonaws.com/role": "role-arn",
												"remove-me":              "true",
											},
										},
									},
								},
							}),
					},
				},
			},
			expected: &annotationsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(ingress, "ing1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Ingress",
								"metadata": map[string]interface{}{
									"name": "ing1",
									"annotations": map[string]interface{}{
										"kubernetes.io/ingress.class": "nginx",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name":        "deploy1",
									"annotations": map[string]interface{}{},
								},
								"spec": map[string]interface{}{
									"template": map[string]interface{}{
										"metadata": map[string]interface{}{
											"annotations": map[string]interface{}{
												"iam.amazonaws.com/role": "role-arn",
											},
										},
									},
								},
							}),
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewAnnotationsTransformer(test.keys)
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config),
					spew.Sdump(test.input.config),
				)
			}

			if !reflect.DeepEqual(test.input.resources, test.expected.resources) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.resources),
					spew.Sdump(test.input.resources),
				)
			}
		})
	}
}
