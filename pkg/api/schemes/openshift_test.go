package schemes

import (
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestAddToScheme(t *testing.T) {
	cases := []struct {
		Name        string
		Groups      []string
		Scheme      func() (*runtime.Scheme, error)
		Validate    func(s *runtime.Scheme, groups []string)
		ExpectError bool
	}{
		{
			Name: "Add openshift api groups",
			Groups: []string{
				"apps.openshift.io",
				"authorization.openshift.io",
				"build.openshift.io",
				"image.openshift.io",
				"route.openshift.io",
				"template.openshift.io",
			},
			Scheme: func() (*runtime.Scheme, error) {
				scheme := runtime.NewScheme()
				err := AddToScheme(scheme)
				if err != nil {
					return nil, err
				}

				return scheme, nil
			},
			Validate: func(s *runtime.Scheme, groups []string) {
				for _, group := range groups {
					if !s.IsGroupRegistered(group) {
						t.Fatalf("Could not api find group: %s", group)
					}
				}
			},
			ExpectError: false,
		},
	}

	for _, tc := range cases {
		scheme, err := tc.Scheme()

		if tc.ExpectError && err == nil {
			t.Fatalf("\"%s\" expected an error but got none", tc.Name)
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("\"%s\" did not expect error but got %s ", tc.Name, err)
		}

		tc.Validate(scheme, tc.Groups)
	}
}
