package plugin

import (
	"fmt"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/rollout/trafficrouting/plugin/rpc"
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	"github.com/argoproj/argo-rollouts/utils/rollout"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
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
		"rollout status":         ro.Status,
		"desiredWeight":          desiredWeight,
		"additionalDestinations": additionalDestinations,
	}).Info("SetWeight called")

	// This checks that we are performing a canary rollout, it is not
	// an error if this is empty. This will be empty on the initial rollout
	if reflect.DeepEqual(ro.Status.Canary, v1alpha1.CanaryStatus{}) {
		r.LogCtx.WithFields(logrus.Fields{"desiredWeight": desiredWeight}).Debug("Rollout does not have a CanaryStatus yet")
		return pluginTypes.RpcError{}
	}

	if rolloutAborted(ro) {
		r.LogCtx.Info("Rollout is aborted, skipping weight update")
		//todo: revert weights
		return pluginTypes.RpcError{}
	}

	if rollout.IsFullyPromoted(ro) {
		r.LogCtx.Info("Rollout is fully promoted, skipping weight update")
		if ro.Status.Canary.Weights.Canary.Weight == 0 {
			r.LogCtx.Info("Rollout is fully promoted and weight is 0, skipping weight update")
			if ro.Status.Canary.Weights.Stable.Weight == 100 {
				rpcError, done := r.updateWeightsOnNats(ro.Name+"-"+ro.Status.CurrentPodHash, 1)
				if done {
					return rpcError
				}
			}
		}
		return pluginTypes.RpcError{}
	}

	if ro.Status.ControllerPause {
		r.LogCtx.Info("Rollout is pausing, skipping weight update")
		return pluginTypes.RpcError{}
	}

	r.LogCtx.WithFields(logrus.Fields{
		"rollout": ro,
	}).Info("SetWeight called deep")

	rpcError, doneCanary := r.updateWeightsOnNats(ro.Name+"-"+ro.Status.CurrentPodHash, desiredWeight)
	if doneCanary {
		return rpcError
	}

	rpcError, doneStable := r.updateWeightsOnNats(ro.Name+"-"+ro.Status.StableRS, 100-desiredWeight)
	if doneStable {
		return rpcError
	}

	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) updateWeightsOnNats(rsName string, desiredWeight int32) (pluginTypes.RpcError, bool) {
	url := fmt.Sprintf("http://host.minikube.internal:8222/debug/weight?weight=%d&desc=%s", desiredWeight, rsName)
	r.LogCtx.WithField("url", url).Info("sending weight request")

	resp, err := http.Get(url)
	if err != nil {
		return pluginTypes.RpcError{ErrorString: fmt.Sprintf("failed to send weight request: %v", err)}, true
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pluginTypes.RpcError{ErrorString: fmt.Sprintf("weight request failed with status: %d", resp.StatusCode)}, true
	}
	return pluginTypes.RpcError{}, false
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

func rolloutAborted(rollout *v1alpha1.Rollout) bool {
	return rollout.Status.Abort
}
