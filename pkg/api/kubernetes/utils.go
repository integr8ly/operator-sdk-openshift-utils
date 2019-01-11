package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strings"

	"encoding/json"

	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/schemes"
)

var (
	scheme      = runtime.NewScheme()
	codecs      = serializer.NewCodecFactory(scheme)
	decoderFunc = decoder
)

func init() {
	schemes.AddToScheme(scheme)
}

func decoder(gv schema.GroupVersion, codecs serializer.CodecFactory) runtime.Decoder {
	codec := codecs.UniversalDecoder(gv)
	return codec
}

func RuntimeObjectFromUnstructured(u *unstructured.Unstructured) (runtime.Object, error) {
	gvk := u.GroupVersionKind()
	decoder := decoderFunc(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on unstructured object: %v", err)
	}

	ro, _, err := decoder.Decode(b, &gvk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}

	return ro, nil
}

func isYaml(filename string) bool {
	for _, suffix := range []string{".yaml", "yaml"} {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}

	return false
}

func LoadKubernetesResourceFromFile(path string) (runtime.Object, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data, err = JsonIfYaml(data, path)
	if err != nil {
		return nil, err
	}

	return LoadKubernetesResource(data)
}

func LoadKubernetesResource(jsonData []byte) (runtime.Object, error) {
	u := unstructured.Unstructured{}

	err := u.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return RuntimeObjectFromUnstructured(&u)
}

func JsonIfYaml(source []byte, filename string) ([]byte, error) {
	if isYaml(filename) {
		return yaml.ToJSON(source)
	}

	return source, nil
}

func UnstructuredFromRuntimeObject(ro runtime.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(ro)
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on runtime object: %v", err)
	}
	var u unstructured.Unstructured
	if err := json.Unmarshal(b, &u.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json into unstructured object: %v", err)
	}
	return &u, nil
}
