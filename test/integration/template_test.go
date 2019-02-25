package integration

import (
	"flag"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/template"
	"github.com/openshift/api/apps/v1"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

var master = new(string)

func init() {
	flag.StringVar(master, "master", "", "openshift master url")
	flag.Parse()
}

func TestTemplateProcessing(t *testing.T) {
	ns := "apicurio"
	params := map[string]string{
		"OPENSHIFT_HOST": "master.host",
	}
	path := "_testdata/template.json"
	var cfg *rest.Config

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to open mock file: %v", err)
	}

	if len(os.Getenv("KUBECONFIG")) > 0 {
		cfg, err = clientcmd.BuildConfigFromFlags(*master, os.Getenv("KUBECONFIG"))
		if err != nil {
			t.Fatal(err)
		}
	}

	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube", "config"))
	if err != nil {
		t.Fatal(err)
	}

	tmpl, err := template.New(cfg, jsonData)
	if err != nil {
		t.Fatal(err)
	}

	err = tmpl.Process(params, ns)
	if err != nil {
		t.Fatal(err)
	}

	objs := tmpl.GetObjects(template.NoFilterFn)
	for _, ro := range objs {
		if ro.GetObjectKind().GroupVersionKind().Kind == "DeploymentConfig" {
			dc := ro.(*v1.DeploymentConfig)
			for _, e := range dc.Spec.Template.Spec.Containers[0].Env {
				if e.Name == "OPENSHIFT_HOST" && e.Value != "master.host" {
					t.Fatalf("Failed to set template param: %v", e)
				}
			}
		}
	}
}
