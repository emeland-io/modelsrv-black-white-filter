package filteredmodel_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFilteredmodel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "filteredmodel Suite")
}
