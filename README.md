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
objects := make([]runtime.Object, 0)
tmpl.CopyObjects(template.NoFilterFn, &objects)
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
    templateData, err := r.box.Find(cr.Spec.Template.Path)
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

    objects := make([]runtime.Object, 0)
    tmpl.CopyObjects(template.NoFilterFn, &objects)

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