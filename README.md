# Operator SDK Openshift Utilities

Library to be used in operator-sdk for openshift specific support.

Features:

* Template inner objects processing into runtime objects
* Register openshift specific types


## Installation

### Dep

Adding the module into your `Gopkg.toml`:

```
[[constraint]]
  name = "github.com/integr8ly/operator-sdk-openshift-utils"
  branch = "master"
```

It can now be installed by running `dep ensure`

## Usage

First of all, you need to register openshift scheme types in your operator (the path of this file is usually `pkg/apis/$group/addtoscheme_$group_$version.go`):


```
package apis

import (
    "github.com/integr8ly/deployment-operator/pkg/apis/integreatly/v1alpha1"
    "github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/schemes"
)

func init() {
    // Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
    AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
    AddToSchemes = append(AddToSchemes, schemes.AddToScheme)
}

```

You will probably need to import these modules:

```
"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/template"
"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/kubernetes"
```

Create a new Tmpl instance:

```go
//r.config is your kubernetes rest.Config struct
tmpl, err := template.New(r.config, jsonData)
if err != nil {
    return err
}
```

You can also create a Template using a reader interface:

```go
type TemplateReader struct {
    data []byte
    readIndex int64
}

func (r *TemplateReader) Read(p []byte) (n int, err error) {
    if r.readIndex >= int64(len(r.data)) {
        err = io.EOF
        return
    }

    n = copy(p, r.data[r.readIndex:])
    r.readIndex += int64(n)
    return
}

b, err := ioutil.ReadFile("_testdata/template.json")
if err != nil {
    panic("Could not read file")
}

reader := &TemplateReader{
    data: b,
}

tmpl := template.FromReader(restConfig, reader)
```

Process the template:

```go
//Paramaters is a map which contains the parameters values that will be filled in the template
err = tmpl.Process(cr.Spec.Template.Parameters, cr.Namespace)
if err != nil {
    return err
}
```

Get the runtime objects:

```go
//copying to an existing slice
objects := make([]runtime.Object, 0)
tmpl.CopyObjects(template.NoFilterFn, &objects)

//retrieving it
objects := tmpl.GetObjects(template.NoFilterFn)
```

Creating runtime objects in the sdk (0.1.1):

```
for _, obj := range objects {
    uo, _ := kubernetes.UnstructuredFromRuntimeObject(obj)
    uo.SetNamespace(cr.Namespace) //namespace needs to be set before creating the project

    err = r.client.Create(context.TODO(), uo.DeepCopyObject())
    if err != nil {
        return err
    }
}
```


Full sample code:

```go
func (r *ReconcileDeployment) DeployTemplate(cr *integreatlyv1alpha1.TDeployment) error {
    var err error
    templateData, err := ioutil.ReadFile(cr.Spec.Template.Path)
    if err != nil {
        return err
    }

    jsonData, err := yaml.ToJSON(templateData)
    if err != nil {
        return err
    }
    log.Printf("%s", string(jsonData[:]))

    tmpl, err := template.New(r.config, jsonData)
    if err != nil {
        return err
    }

    err = tmpl.Process(cr.Spec.Template.Parameters, cr.Namespace)
    if err != nil {
        return err
    }

    /*objects := make([]runtime.Object, 0)
    tmpl.CopyObjects(template.NoFilterFn, &objects)*/
    
    objects := tmpl.GetObjects(template.NoFilterFn)

    for _, obj := range objects {
        uo, _ := kubernetes.UnstructuredFromRuntimeObject(obj)
        uo.SetNamespace(cr.Namespace)

        err = r.client.Create(context.TODO(), uo.DeepCopyObject())
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Development

Unit tests:

```sh
make test/unit
```

Smoke tests (checks syntax + unit tests):

```sh
make test/smoke
```

Fixing code formatting:

```sh
make code/fix
```