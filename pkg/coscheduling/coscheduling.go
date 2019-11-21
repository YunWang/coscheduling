package coscheduling

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	gangplugin "github.com/YunWang/gangplugin/pkg/api/v1"
	v1 "k8s.io/api/core/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

type CoSchedulingPlugin struct {
	FrameworkHandler framework.FrameworkHandle
	client.Client
}

const Name  = "CoSchedulingPermitPlugin"

func (cs *CoSchedulingPlugin) Name() string{
	return Name
}

func (cs *CoSchedulingPlugin) Permit(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (*framework.Status, time.Duration){

	//get gang name from pod's annotation
	gangName,exist:= p.Annotations[gangplugin.GangKey]
	if !exist {
		//return true directly
		return framework.NewStatus(framework.Success,""),0
	}
	//get gang according gangName and pod.Namespace
	targetGang:=&gangplugin.Gang{}
	err := cs.Get(ctx,types.NamespacedName{Name:gangName,Namespace:p.Namespace},targetGang)
	if err!=nil{
		//gang not found and return true directly
		return framework.NewStatus(framework.Success,""),0
	}
	//minGang defines minimal number of pods to run
	if targetGang.Spec.MinGang<=1 {
		return framework.NewStatus(framework.Success,""),0
	}

	//iterate waiting pods and count gang pod number
	count := int32(1) //count gang pod number
	search := func(p framework.WaitingPod) {
		// TODO: add more checks for these pods, e.g. whether it has been deleted
		pod:=p.GetPod()
		//check whether pod still exist
		_,exist := cs.FrameworkHandler.SharedInformerFactory().Core().V1().Pods().Lister().Pods(pod.Namespace).Get(pod.Name)
		if exist != nil{
			//if pod has been deleted, then delete it from WaitingPod
			cs.FrameworkHandler.RejectWaitingPod(pod.UID)
		}
		if pod.Annotations[gangplugin.GangKey] == gangName {
			count++
		}
	}
	cs.FrameworkHandler.IterateOverWaitingPods(search)

}

func New(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	return &CoSchedulingPlugin{},nil
}
