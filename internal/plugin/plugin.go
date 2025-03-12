package plugin

import (
	"fmt"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/rollout/trafficrouting/plugin/rpc"
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	"github.com/sirupsen/logrus"
)

var _ rpc.TrafficRouterPlugin = &RpcPlugin{}

type RpcPlugin struct {
	LogCtx *logrus.Entry
}

func (p *RpcPlugin) InitPlugin() pluginTypes.RpcError {
	p.LogCtx.Info("InitPlugin")

	/*loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here
	configOverrides := &clientcmd.ConfigOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	kubeClient, _ := kubernetes.NewForConfig(config)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		kubeClient,
		5*time.Minute,
		kubeinformers.WithNamespace(metav1.NamespaceAll))

	//cache.WaitForCacheSync(context.Background().Done(), p.ingressWrapper.Informer().HasSynced)
	*/
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetWeight(ro *v1alpha1.Rollout, desiredWeight int32, additionalDestinations []v1alpha1.WeightDestination) pluginTypes.RpcError {
	r.LogCtx.WithFields(logrus.Fields{
		"rollout":                ro.Name,
		"desiredWeight":          desiredWeight,
		"additionalDestinations": additionalDestinations,
	}).Info("SetWeight called")

	//s := v1alpha1.NginxTrafficRouting{}
	//err := json.Unmarshal(ro.Spec.Strategy.Canary.TrafficRouting.Plugins["surovskyi/sample-nats"], &s)
	//if err != nil {
	//	return pluginTypes.RpcError{ErrorString: "could not unmarshal nats config"}
	//}

	//stableIngressName := s.StableIngress
	//canaryIngressName := getCanaryIngressName(ro, stableIngressName)

	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetHeaderRoute(ro *v1alpha1.Rollout, headerRouting *v1alpha1.SetHeaderRoute) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) VerifyWeight(ro *v1alpha1.Rollout, desiredWeight int32, additionalDestinations []v1alpha1.WeightDestination) (pluginTypes.RpcVerified, pluginTypes.RpcError) {
	return pluginTypes.NotImplemented, pluginTypes.RpcError{}
}

// UpdateHash informs a traffic routing reconciler about new canary/stable pod hashes
func (r *RpcPlugin) UpdateHash(ro *v1alpha1.Rollout, canaryHash, stableHash string, additionalDestinations []v1alpha1.WeightDestination) pluginTypes.RpcError {
	r.LogCtx.WithFields(logrus.Fields{
		"rollout":                ro.Name,
		"canaryHash":             canaryHash,
		"stableHash":             stableHash,
		"additionalDestinations": additionalDestinations,
	}).Info("UpdateHash called")

	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetMirrorRoute(ro *v1alpha1.Rollout, setMirrorRoute *v1alpha1.SetMirrorRoute) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) RemoveManagedRoutes(ro *v1alpha1.Rollout) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) Type() string {
	return "plugin-nats"
}

func getCanaryIngressName(rollout *v1alpha1.Rollout, stableIngress string) string {
	// names limited to 253 characters
	if rollout.Spec.Strategy.Canary != nil &&
		rollout.Spec.Strategy.Canary.TrafficRouting != nil &&
		rollout.Spec.Strategy.Canary.TrafficRouting.Plugins != nil &&
		rollout.Spec.Strategy.Canary.TrafficRouting.Plugins["argoproj/sample-nats"] != nil {

		prefix := fmt.Sprintf("%s-%s", rollout.GetName(), stableIngress)
		if len(prefix) > 253-len("-canary") {
			// trim prefix
			prefix = prefix[0 : 253-len("-canary")]
		}
		return fmt.Sprintf("%s%s", prefix, "-canary")
	}
	return ""
}
