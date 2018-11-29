package schemes

import (
	"k8s.io/apimachinery/pkg/runtime"

	apps "github.com/openshift/api/apps/v1"
	authorization "github.com/openshift/api/authorization/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	template "github.com/openshift/api/template/v1"
)

var AddToSchemes runtime.SchemeBuilder

func init() {
	add()
}

func add() {
	AddToSchemes = append(AddToSchemes,
		apps.Install,
		authorization.Install,
		build.Install,
		image.Install,
		route.Install,
		template.Install,
	)
}

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
