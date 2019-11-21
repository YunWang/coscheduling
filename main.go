package coscheduling

import (
	"fmt"
	"os"

	scheduler "k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/wangyun/coscheduling/pkg/coscheduling"
)

func main() {
	command := scheduler.NewSchedulerCommand(
		scheduler.WithPlugin(coscheduling.Name, coscheduling.New))

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}