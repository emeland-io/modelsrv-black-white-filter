package filterconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFilterconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "filterconfig Suite")
}
