package kubernetes

import (
	"io/ioutil"
	"testing"
)

func TestLoadKubernetesResourceFromFile(t *testing.T) {

	cases := []struct {
		Name        string
		FilePath    string
		Resource    func(path string) error
		ExpectError bool
	}{
		{
			Name:     "Should open and parse yaml file",
			FilePath: "_testdata/test-template.yaml",
			Resource: func(path string) error {
				_, err := LoadKubernetesResourceFromFile(path)

				return err
			},
			ExpectError: false,
		},
		{
			Name:     "Should open and parse json file",
			FilePath: "_testdata/test-template.json",
			Resource: func(path string) error {
				_, err := LoadKubernetesResourceFromFile(path)

				return err
			},
			ExpectError: false,
		},
		{
			Name:     "Should fail to find template file",
			FilePath: "_testdata/test-template",
			Resource: func(path string) error {
				_, err := LoadKubernetesResourceFromFile(path)

				return err
			},
			ExpectError: true,
		},
		{
			Name:     "Should fail to find template file",
			FilePath: "_testdata/test-invalid-template.json",
			Resource: func(path string) error {
				_, err := LoadKubernetesResourceFromFile(path)

				return err
			},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		err := tc.Resource(tc.FilePath)

		if tc.ExpectError && err == nil {
			t.Fatalf("expected an error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("did not expect error but got %s ", err)
		}
	}
}

func TestJsonIfYaml(t *testing.T) {
	cases := []struct {
		Name        string
		FilePath    string
		Content     func(path string) []byte
		ExpectError bool
	}{
		{
			Name:     "Should parse yaml file",
			FilePath: "_testdata/test-template.yaml",
			Content: func(path string) []byte {
				rawData, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatalf("Could not find template file %v", err)
				}

				return rawData
			},
			ExpectError: false,
		},
		{
			Name:     "Should parse json file",
			FilePath: "_testdata/test-template.json",
			Content: func(path string) []byte {
				rawData, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatalf("Could not find template file %v", err)
				}

				return rawData
			},
			ExpectError: false,
		},
		{
			Name:     "Should fail to find template file",
			FilePath: "_testdata/test-invalid-template.jso",
			Content: func(path string) []byte {
				rawData, err := ioutil.ReadFile(path)
				if err != nil {
					return nil
				}

				return rawData
			},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		content := tc.Content(tc.FilePath)

		if tc.ExpectError {
			if len(content) > 0 {
				t.Fatalf("expected an error but got none")
			}

			if len(content) == 0 {
				continue
			}
		}

		_, err := JsonIfYaml(content, tc.FilePath)

		if tc.ExpectError && err == nil {
			t.Fatalf("expected an error but got none")
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("did not expect error but got %s ", err)
		}
	}
}
