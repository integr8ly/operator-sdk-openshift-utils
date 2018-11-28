package template

import (
	v1template "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var (
	TmplDefaultOpts = TmplOpt{
		ApiVersion:  "v1",
		ApiMimetype: "application/json",
		ApiPath:     "/apis",
		ApiGroup:    "template.openshift.io",
		ApiResource: "processedtemplates",
	}
)

type Tmpl struct {
	RestClient rest.Interface
	Source     *v1template.Template
	Raw        []byte
	Objects    []runtime.Object
}

type FilterFn func(obj *runtime.Object) error

type TmplOpt struct {
	ApiKind     string
	ApiVersion  string
	ApiPath     string
	ApiGroup    string
	ApiMimetype string
	ApiResource string
}
