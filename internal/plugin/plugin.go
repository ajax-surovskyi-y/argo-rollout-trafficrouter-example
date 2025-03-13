package plugin

import (
	"fmt"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/rollout/trafficrouting/plugin/rpc"
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	"github.com/argoproj/argo-rollouts/utils/rollout"
	"github.com/sirupsen/logrus"
	"net/http"
)

var _ rpc.TrafficRouterPlugin = &RpcPlugin{}

type RpcPlugin struct {
	LogCtx *logrus.Entry
}

func (p *RpcPlugin) InitPlugin() pluginTypes.RpcError {
	p.LogCtx.Info("InitPlugin")
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetWeight(ro *v1alpha1.Rollout, desiredWeight int32, additionalDestinations []v1alpha1.WeightDestination) pluginTypes.RpcError {
	r.LogCtx.WithFields(logrus.Fields{
		"rollout":                ro,
		"desiredWeight":          desiredWeight,
		"additionalDestinations": additionalDestinations,
	}).Info("SetWeight called")

	if rollout.IsFullyPromoted(ro) {
		r.LogCtx.Info("Rollout is fully promoted, skipping weight update")
		return pluginTypes.RpcError{}
	}

	if !rollout.IsUnpausing(ro) {
		r.LogCtx.Info("Rollout is pausing, skipping weight update")
		return pluginTypes.RpcError{}
	}

	url := fmt.Sprintf("http://host.minikube.internal:8222/debug/weight?weight=%d&desc=%s", desiredWeight, ro.Name+"-"+ro.Status.CurrentPodHash)
	r.LogCtx.WithField("url", url).Info("sending weight request")

	resp, err := http.Get(url)
	if err != nil {
		return pluginTypes.RpcError{ErrorString: fmt.Sprintf("failed to send weight request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pluginTypes.RpcError{ErrorString: fmt.Sprintf("weight request failed with status: %d", resp.StatusCode)}
	}

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
