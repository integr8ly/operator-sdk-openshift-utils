package template

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/openshift/api/template/v1"
	"io"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"net/http"
	"testing"
)

func objBody(object interface{}) io.ReadCloser {
	output, err := json.MarshalIndent(object, "", "")
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(output)))
}

func TestNew(t *testing.T) {
	cases := []struct {
		Name        string
		Template    func() (*Tmpl, error)
		ExpectError bool
	}{
		{
			Name: "Should create a new template",
			Template: func() (*Tmpl, error) {
				return New(&rest.Config{}, []byte{})
			},
			ExpectError: false,
		},
	}

	for _, tc := range cases {
		_, err := tc.Template()

		if tc.ExpectError && err == nil {
			t.Fatal("Expected error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("Test failed: %v", err)
		}
	}
}

func TestTmpl_Bootstrap(t *testing.T) {
	cases := []struct {
		Name        string
		Template    *Tmpl
		Config      *rest.Config
		Opts        TmplOpt
		ExpectError bool
	}{
		{
			Name:        "Should bootstrap template",
			Template:    &Tmpl{},
			Config:      &rest.Config{},
			Opts:        TmplDefaultOpts,
			ExpectError: false,
		},
		{
			Name:     "Should fail to bootstrap template",
			Template: &Tmpl{},
			Config:   &rest.Config{},
			Opts: TmplOpt{
				ApiVersion:  "v0",
				ApiMimetype: "text/xml",
				ApiPath:     "/soa",
				ApiGroup:    "soap.openshift.io",
				ApiResource: "wsdl",
			},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		err := tc.Template.Bootstrap(tc.Config, tc.Opts)

		if tc.ExpectError && err == nil {
			t.Fatal("Expected error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("Test failed: %v", err)
		}

	}
}

func TestTmpl_Process(t *testing.T) {
	cases := []struct {
		Name        string
		Template    func(path string) *Tmpl
		Path        string
		Params      map[string]string
		Namespace   string
		ExpectError bool
	}{
		{
			Name: "Should process template",
			Template: func(path string) *Tmpl {
				b, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to open mock file: %v", err)
				}

				serverVersions := []string{"/v1", "/templates"}
				body := ioutil.NopCloser(bytes.NewReader(b))
				client := &fake.RESTClient{
					NegotiatedSerializer: kubescheme.Codecs,
					Resp: &http.Response{
						StatusCode: 201,
						Body:       objBody(&metav1.APIVersions{Versions: serverVersions}),
					},
					Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
						header := http.Header{}
						header.Set("Content-Type", "application/json")
						return &http.Response{StatusCode: 201, Header: header, Body: body}, nil
					}),
				}

				return &Tmpl{
					Raw:        b,
					Source:     &v1.Template{},
					RestClient: client,
				}
			},
			Path:        "_testdata/template.json",
			Params:      map[string]string{},
			Namespace:   "test",
			ExpectError: false,
		},
	}

	for _, tc := range cases {
		tmpl := tc.Template(tc.Path)
		err := tmpl.Process(tc.Params, tc.Namespace)

		if tc.ExpectError && err == nil {
			t.Fatal("Expected error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("Test failed: %v", err)
		}
	}
}

func TestTmpl_FillObjects(t *testing.T) {
	cases := []struct {
		Name        string
		Template    *Tmpl
		Extensions  func() []runtime.RawExtension
		Validate    func(tmpl *Tmpl)
		ExpectError bool
	}{
		{
			Name:     "Should fill template objects",
			Template: &Tmpl{},
			Extensions: func() []runtime.RawExtension {
				exts := make([]runtime.RawExtension, 0)
				for _, path := range []string{"pod.json", "service.json", "route.json"} {
					b, err := ioutil.ReadFile("_testdata/" + path)
					if err != nil {
						t.Fatalf("Failed to open mock file: %v", err)
					}
					ext := runtime.RawExtension{
						Raw: b,
					}
					exts = append(exts, ext)
				}

				return exts
			},
			Validate: func(tmpl *Tmpl) {
				if len(tmpl.Objects) != 3 {
					t.Fatalf("Failed to fill template objects: %v", tmpl.Objects)
				}
			},
			ExpectError: false,
		},
		{
			Name:     "Should fail to fill object",
			Template: &Tmpl{},
			Extensions: func() []runtime.RawExtension {
				exts := make([]runtime.RawExtension, 0)
				for _, path := range []string{"pod.json", "custom-object.json"} {
					b, err := ioutil.ReadFile("_testdata/" + path)
					if err != nil {
						t.Fatalf("Failed to open mock file: %v", err)
					}
					ext := runtime.RawExtension{
						Raw: b,
					}
					exts = append(exts, ext)
				}

				return exts
			},
			Validate:    func(tmpl *Tmpl) {},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		err := tc.Template.FillObjects(tc.Extensions())

		if tc.ExpectError && err == nil {
			t.Fatal("Expected error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc.Validate(tc.Template)
	}
}

func TestTmpl_CopyObjects(t *testing.T) {
	cases := []struct {
		Name       string
		Objects    []runtime.Object
		CallbackFn FilterFn
		Validate   func(tmpl *Tmpl, objects []runtime.Object)
	}{
		{
			Name: "Should copy all objects",
			Objects: []runtime.Object{
				&runtime.Unknown{},
				&runtime.Unknown{},
				&runtime.Unknown{},
			},
			CallbackFn: NoFilterFn,
			Validate: func(tmpl *Tmpl, objects []runtime.Object) {
				if len(tmpl.Objects) != len(objects) {
					t.Fatalf("Failed to copy objects: %v", objects)
				}
			},
		},
		{
			Name: "Should copy json objects only",
			Objects: []runtime.Object{
				&runtime.Unknown{
					ContentType: "application/json",
				},
				&runtime.Unknown{
					ContentType: "text/xml",
				},
				&runtime.Unknown{
					ContentType: "application/json",
				},
			},
			CallbackFn: func(obj *runtime.Object) error {
				o := *obj

				u := o.(*runtime.Unknown)
				if u.ContentType == "text/xml" {
					return errors.New("xml not allowed")
				}

				return nil
			},
			Validate: func(tmpl *Tmpl, objects []runtime.Object) {
				if len(objects) != 2 {
					t.Fatalf("Failed to copy objects: %v", objects)
				}

				for _, obj := range objects {
					u := obj.(*runtime.Unknown)
					if u.ContentType != "application/json" {
						t.Fatalf("Invalid object copied: %v", u)
					}
				}
			},
		},
	}

	for _, tc := range cases {
		objs := make([]runtime.Object, 0)
		tmpl := &Tmpl{
			Objects: tc.Objects,
		}
		tmpl.CopyObjects(tc.CallbackFn, &objs)

		tc.Validate(tmpl, objs)
	}
}

func TestTmpl_FillParams(t *testing.T) {
	cases := []struct {
		Name     string
		Params   map[string]string
		Template *Tmpl
		Validate func(tmpl *Tmpl, params map[string]string)
	}{
		{
			Name: "Should fill template params",
			Params: map[string]string{
				"p1": "param 1",
				"p2": "param2",
			},
			Template: &Tmpl{
				Source: &v1.Template{
					Parameters: []v1.Parameter{
						{
							Name: "p1",
						},
						{
							Name: "p2",
						},
					},
				},
			},
			Validate: func(tmpl *Tmpl, params map[string]string) {
				for _, param := range tmpl.Source.Parameters {
					if value, ok := params[param.Name]; ok {
						if value != param.Value {
							t.Fatalf("Value differs [%s] = %s", value, param.Value)
						}
					}
				}
			},
		},
		{
			Name:   "Template should have 0 parameters",
			Params: map[string]string{},
			Template: &Tmpl{
				Source: &v1.Template{
					Parameters: []v1.Parameter{},
				},
			},
			Validate: func(tmpl *Tmpl, params map[string]string) {
				if len(tmpl.Source.Parameters) > 0 {
					t.Fatal("Template parameters property should be empty")
				}
			},
		},
	}

	for _, tc := range cases {
		tc.Template.FillParams(tc.Params)
		tc.Validate(tc.Template, tc.Params)
	}
}
