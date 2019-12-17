package fake

import "github.com/jthomperoo/custom-pod-autoscaler/config"

// Execute (fake) allows inserting logic into an executer for testing
type Execute struct {
	ExecuteWithValueReactor func(method *config.Method, value string) (string, error)
	GetTypeReactor          func() string
}

// ExecuteWithValue calls the fake Execute reactor method provided
func (f *Execute) ExecuteWithValue(method *config.Method, value string) (string, error) {
	return f.ExecuteWithValueReactor(method, value)
}

// GetType calls the fake Execute reactor method provided
func (f *Execute) GetType() string {
	return f.GetTypeReactor()
}
