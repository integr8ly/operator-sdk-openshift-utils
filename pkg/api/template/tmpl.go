package template

import (
	"encoding/json"

	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/kubernetes"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/schemes"
	v1template "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func New(restConfig *rest.Config, data []byte) (*Tmpl, error) {
	tmpl := &Tmpl{
		Source: &v1template.Template{},
		Raw:    data,
	}

	err := tmpl.Bootstrap(restConfig, TmplDefaultOpts)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func NoFilterFn(obj *runtime.Object) error {
	return nil
}

func (t *Tmpl) Bootstrap(restConfig *rest.Config, opts TmplOpt) error {
	config := rest.CopyConfig(restConfig)
	config.GroupVersion = &schema.GroupVersion{
		Group:   opts.ApiGroup,
		Version: opts.ApiVersion,
	}
	config.APIPath = opts.ApiPath
	config.AcceptContentTypes = opts.ApiMimetype
	config.ContentType = opts.ApiMimetype

	config.NegotiatedSerializer = schemes.BasicNegotiatedSerializer{}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return err
	}

	t.RestClient = restClient

	return nil
}

func (t *Tmpl) Process(params map[string]string, ns string) error {
	resource, err := json.Marshal(t.Source)
	if err != nil {
		return err
	}

	jsonData, err := kubernetes.LoadKubernetesResource(t.Raw)
	if err != nil {
		return err
	}

	t.Source = jsonData.(*v1template.Template)

	result := t.RestClient.
		Post().
		Namespace(ns).
		Body(resource).
		Resource("processedtemplates").
		Do()

	if result.Error() != nil {
		return result.Error()
	}

	data, err := result.Raw()
	if err != nil {
		return err
	}

	templateObject, err := kubernetes.LoadKubernetesResource(data)
	if err != nil {
		return err
	}

	t.Source = templateObject.(*v1template.Template)
	t.FillParams(params)

	err = t.FillObjects(t.Source.Objects)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tmpl) FillObjects(rawObjects []runtime.RawExtension) error {
	for _, rawObject := range rawObjects {
		obj, err := kubernetes.LoadKubernetesResource(rawObject.Raw)
		if err != nil {
			return err
		}

		t.Objects = append(t.Objects, obj)
	}

	return nil
}

func (t *Tmpl) FillParams(params map[string]string) {
	for i, param := range t.Source.Parameters {
		if value, ok := params[param.Name]; ok {
			t.Source.Parameters[i].Value = value
		}
	}
}

func (t *Tmpl) CopyObjects(filter FilterFn, objects *[]runtime.Object) {
	for _, obj := range t.Objects {
		err := filter(&obj)
		if err == nil {
			*objects = append(*objects, obj.DeepCopyObject())
		}
	}
}
