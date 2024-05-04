package transformers

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/layertwo/helm-convert/pkg/types"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	kimage "sigs.k8s.io/kustomize/pkg/image"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type imageTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestImageRun(t *testing.T) {
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	for _, test := range []struct {
		name     string
		input    *imageTransformerArgs
		expected *imageTransformerArgs
	}{
		{
			name: "it should retrieve images",
			input: &imageTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"spec": map[string]interface{}{
										"template": map[string]interface{}{
											"spec": map[string]interface{}{
												"initContainers": []interface{}{
													map[string]interface{}{
														"name":  "busybox",
														"image": "busybox",
													},
												},
												"containers": []interface{}{
													map[string]interface{}{
														"name":  "nginx",
														"image": "nginx:1.7.9",
													},
													map[string]interface{}{
														"name":  "alpine",
														"image": "alpine@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
													},
													map[string]interface{}{
														"name":  "centos",
														"image": "myregistry:5000/namespace/centos:1.2.3",
													},
												},
											},
										},
									},
								},
							}),
						resid.NewResId(deploy, "deploy2"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"spec": map[string]interface{}{
										"template": map[string]interface{}{
											"spec": map[string]interface{}{
												"containers": []interface{}{
													map[string]interface{}{
														"name":  "nginx",
														"image": "nginx:1.7.9",
													},
												},
											},
										},
									},
								},
							}),
					},
				},
			},
			expected: &imageTransformerArgs{
				config: &ktypes.Kustomization{
					Images: []kimage.Image{
						{Name: "alpine", Digest: "sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
						{Name: "busybox"},
						{Name: "myregistry:5000/namespace/centos", NewTag: "1.2.3"},
						{Name: "nginx", NewTag: "1.7.9"},
					},
				},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"spec": map[string]interface{}{
										"template": map[string]interface{}{
											"spec": map[string]interface{}{
												"initContainers": []interface{}{
													map[string]interface{}{
														"name":  "busybox",
														"image": "busybox",
													},
												},
												"containers": []interface{}{
													map[string]interface{}{
														"name":  "nginx",
														"image": "nginx:1.7.9",
													},
													map[string]interface{}{
														"name":  "alpine",
														"image": "alpine@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
													},
													map[string]interface{}{
														"name":  "centos",
														"image": "myregistry:5000/namespace/centos:1.2.3",
													},
												},
											},
										},
									},
								},
							}),
						resid.NewResId(deploy, "deploy2"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"spec": map[string]interface{}{
										"template": map[string]interface{}{
											"spec": map[string]interface{}{
												"containers": []interface{}{
													map[string]interface{}{
														"name":  "nginx",
														"image": "nginx:1.7.9",
													},
												},
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
		t.Run(test.name, func(t *testing.T) {
			lt := NewImageTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := pretty.Compare(test.input.config, test.expected.config); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}

			if diff := pretty.Compare(test.input.resources, test.expected.resources); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}
		})
	}
}

func TestCreateKImage(t *testing.T) {
	imagePath := "myregistry:5000/namespace/busybox:1.2.3"
	image := createKImage(imagePath)
	if image.Name != "myregistry:5000/namespace/busybox" || image.NewTag != "1.2.3" {
		t.Fatalf("Parsed imageName: %s newTag %s from %s", image.Name, image.NewTag, imagePath)
	}

	imagePath = "myregistry:5000/namespace/busybox@sha256" +
		":24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"
	image = createKImage(imagePath)
	if image.Name != "myregistry:5000/namespace/busybox" || image.Digest != "sha256:"+
		"24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3" {
		t.Fatalf("Parsed imageName: %s digest %s from %s", image.Name, image.Digest, imagePath)
	}

	imagePath = "busybox"
	image = createKImage(imagePath)
	if image.Name != "busybox" {
		t.Fatalf("Parsed imageName: %s newTag %s from %s", image.Name, image.NewTag, imagePath)
	}

	imagePath = "busybox:1.2.3"
	image = createKImage(imagePath)
	if image.Name != "busybox" || image.NewTag != "1.2.3" {
		t.Fatalf("Parsed imageName: %s newTag %s from %s", image.Name, image.NewTag, imagePath)
	}
}
