// +build !unittests

package setup_test

import (
	"flag"
	"log"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	ptpv1 "github.com/openshift/ptp-operator/pkg/apis/ptp/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	_ "github.com/openshift-kni/performance-addon-operators/functests/0_config" // this is needed otherwise the performance test won't be executed
	ginkgo_reporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	testclient "github.com/openshift-kni/cnf-features-deploy/functests/utils/client"
	"github.com/openshift-kni/cnf-features-deploy/functests/utils/k8sreporter"
)

// TODO: we should refactor tests to use client from controller-runtime package
// see - https://github.com/openshift/cluster-api-actuator-pkg/blob/master/pkg/e2e/framework/framework.go

var junitPath *string
var reportPath *string

func init() {
	junitPath = flag.String("junit", "", "the path for the junit format report")
	reportPath = flag.String("report", "", "the path of the report file containing details for failed tests")
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)

	rr := []Reporter{}
	if ginkgo_reporters.Polarion.Run {
		rr = append(rr, &ginkgo_reporters.Polarion)
	}

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, "setup_junit.xml")
		rr = append(rr, reporters.NewJUnitReporter(junitFile))
	}
	if *reportPath != "" {
		reportFile := path.Join(*reportPath, "setup_failure_report.log")
		reporter, output, err := newTestsReporter(reportFile)
		if err != nil {
			log.Fatalf("Failed to create log reporter %s", err)
		}
		defer output.Close()
		rr = append(rr, reporter)
	}

	RunSpecsWithDefaultAndCustomReporters(t, "CNF Features e2e setup", rr)
}

var _ = BeforeSuite(func() {
	Expect(testclient.Client).NotTo(BeNil(), "Verify the KUBECONFIG environment variable")
})

var _ = AfterSuite(func() {

})

func newTestsReporter(reportPath string) (*k8sreporter.KubernetesReporter, *os.File, error) {
	addToScheme := func(s *runtime.Scheme) {
		ptpv1.AddToScheme(s)
		mcfgv1.AddToScheme(s)
	}

	filterPods := func(pod *v1.Pod) bool {
		if pod.Namespace == "sctptest" {
			return false
		}
		if pod.Namespace == "openshift-ptp" {
			return false
		}
		if pod.Namespace == "openshift-performance-addon" {
			return false
		}
		return true
	}

	f, err := os.OpenFile(reportPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	crs := []k8sreporter.CRData{
		k8sreporter.CRData{
			Cr: &mcfgv1.MachineConfigPoolList{},
		},
		k8sreporter.CRData{
			Cr: &ptpv1.PtpConfigList{},
		},
	}
	res, err := k8sreporter.New("", addToScheme, filterPods, f, crs...)
	if err != nil {
		return nil, nil, err
	}
	return res, f, nil
}
