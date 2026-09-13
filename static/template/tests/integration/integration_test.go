package integration_test

import (
	"testing"

	"gosalusa.com/static/template/test"
)

func TestIntegration(t *testing.T) {
	test.Kernel(t).
		GetJSON("/api/user").
		AssertStatus2XX().
		AssertJSONString(`{
			"users": []
		}`)
}
