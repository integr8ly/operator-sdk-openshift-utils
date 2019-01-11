package template

import (
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/kubernetes"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/schemes"
	v1template "github.com/openshift/api/template/v1"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func New(restConfig *rest.Config, data []byte) (*Tmpl, error) {
	tmpl := &Tmpl{
		Raw: data,
	}

	res, err := kubernetes.LoadKubernetesResource(tmpl.Raw)
	if err != nil {
		return nil, err
	}
	tmpl.Source = res.(*v1template.Template)

	err = tmpl.Bootstrap(restConfig, TmplDefaultOpts)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func FromReader(restConfig *rest.Config, reader io.Reader) (*Tmpl, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return New(restConfig, data)
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
	var err error

	t.fillParams(params)

	uo, err := kubernetes.UnstructuredFromRuntimeObject(t.Source.DeepCopyObject())
	if err != nil {
		return err
	}

	jsonData, err := uo.MarshalJSON()
	if err != nil {
		return err
	}

	result := t.RestClient.
		Post().
		Namespace(ns).
		Body(jsonData).
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

	err = t.fillObjects(t.Source.Objects)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tmpl) fillObjects(rawObjects []runtime.RawExtension) error {
	for _, rawObject := range rawObjects {
		obj, err := kubernetes.LoadKubernetesResource(rawObject.Raw)
		if err != nil {
			return err
		}

		t.Objects = append(t.Objects, obj)
	}

	return nil
}

func (t *Tmpl) fillParams(params map[string]string) {
	for i, param := range t.Source.Parameters {
		if value, ok := params[param.Name]; ok {
			t.Source.Parameters[i].Value = value
		}
	}
}

func (t *Tmpl) GetObjects(filter FilterFn) []runtime.Object {
	objects := make([]runtime.Object, 0)

	for _, obj := range t.Objects {
		err := filter(&obj)
		if err == nil {
			objects = append(objects, obj.DeepCopyObject())
		}
	}

	return objects
}

func (t *Tmpl) CopyObjects(filter FilterFn, objects *[]runtime.Object) {
	*objects = t.GetObjects(filter)
}
