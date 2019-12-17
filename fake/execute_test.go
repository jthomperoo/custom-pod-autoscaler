package fake_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
)

func TestExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description             string
		expected                string
		expectedErr             error
		method                  *config.Method
		value                   string
		executeWithValueReactor func(method *config.Method, value string) (string, error)
	}{
		{
			"Return error",
			"",
			errors.New("execute error"),
			&config.Method{},
			"test",
			func(method *config.Method, value string) (string, error) {
				return "", errors.New("execute error")
			},
		},
		{
			"Return without error",
			"test",
			nil,
			&config.Method{},
			"test",
			func(method *config.Method, value string) (string, error) {
				return "test", nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &fake.Execute{
				ExecuteWithValueReactor: test.executeWithValueReactor,
			}
			result, err := execute.ExecuteWithValue(test.method, test.value)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestExecute_GetType(t *testing.T) {
	var tests = []struct {
		description    string
		expected       string
		getTypeReactor func() string
	}{
		{
			"Return type",
			"test",
			func() string {
				return "test"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &fake.Execute{
				GetTypeReactor: test.getTypeReactor,
			}
			result := execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
