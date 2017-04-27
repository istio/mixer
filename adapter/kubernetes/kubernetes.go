// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package kubernetes provides functionality to adapt mixer behavior to the
// kubernetes environment. Primarily, it is used to generate values as part
// of Mixer's attribute generation preprocessing phase. These values will be
// transformed into attributes that can be used for subsequent config
// resolution and adapter dispatch and execution.
package kubernetes

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"istio.io/mixer/adapter/kubernetes/config"
	"istio.io/mixer/pkg/adapter"
)

type (
	builder struct {
		adapter.DefaultBuilder

		stopChan             chan struct{}
		pods                 cacheController
		once                 sync.Once
		newCacheControllerFn controllerFactoryFn
	}
	kubegen struct {
		log    adapter.Logger
		pods   cacheController
		params config.Params
	}

	// used strictly for testing purposes
	controllerFactoryFn func(kubeconfigPath string, refreshDuration time.Duration, env adapter.Env) (cacheController, error)
)

const (
	// adapter vals
	name = "kubernetes"
	desc = "Provides platform specific functionality for the kubernetes environment"

	// parsing
	kubePrefix       = "kubernetes://"
	defaultNamespace = "default"

	// input/output naming
	sourceUID         = "sourceUID"
	targetUID         = "targetUID"
	originUID         = "originUID"
	sourcePrefix      = "source"
	targetPrefix      = "target"
	originPrefix      = "origin"
	labelsVal         = "Labels"
	podNameVal        = "PodName"
	podIPVal          = "PodIP"
	hostIPVal         = "HostIP"
	namespaceVal      = "Namespace"
	serviceAccountVal = "ServiceAccountName"
	serviceVal        = "Service"

	// value extraction
	targetService   = "targetService"
	clusterDomain   = "svc.cluster.local"
	podServiceLabel = "app"

	// cache invaliation
	// TODO: determine a reasonable default
	defaultRefreshPeriod = 5 * time.Minute
)

var (
	conf = &config.Params{
		KubeconfigPath:          "",
		CacheRefreshDuration:    defaultRefreshPeriod,
		SourceUidInputName:      sourceUID,
		TargetUidInputName:      targetUID,
		OriginUidInputName:      originUID,
		TargetServiceInputName:  targetService,
		ClusterDomainName:       clusterDomain,
		PodLabelForService:      podServiceLabel,
		SourcePrefix:            sourcePrefix,
		TargetPrefix:            targetPrefix,
		OriginPrefix:            originPrefix,
		LabelsValueName:         labelsVal,
		PodNameValueName:        podNameVal,
		PodIpValueName:          podIPVal,
		HostIpValueName:         hostIPVal,
		NamespaceValueName:      namespaceVal,
		ServiceAccountValueName: serviceAccountVal,
		ServiceValueName:        serviceVal,
	}
)

// Register records the builders exposed by this adapter.
func Register(r adapter.Registrar) {
	r.RegisterAttributesGeneratorBuilder(newBuilder(newCacheFromConfig))
}

func newBuilder(cacheFactory controllerFactoryFn) *builder {
	stopChan := make(chan struct{})
	return &builder{adapter.NewDefaultBuilder(name, desc, conf), stopChan, nil, sync.Once{}, cacheFactory}
}

func (b *builder) Close() error {
	close(b.stopChan)
	return nil
}

func (*builder) ValidateConfig(c adapter.Config) (ce *adapter.ConfigErrors) {
	params := c.(*config.Params)
	if len(params.SourceUidInputName) == 0 {
		ce = ce.Appendf("sourceUidInputName", "field must be populated")
	}
	if len(params.TargetUidInputName) == 0 {
		ce = ce.Appendf("targetUidInputName", "field must be populated")
	}
	if len(params.OriginUidInputName) == 0 {
		ce = ce.Appendf("originUidInputName", "field must be populated")
	}
	if len(params.SourcePrefix) == 0 {
		ce = ce.Appendf("sourcePrefix", "field must be populated")
	}
	if len(params.TargetPrefix) == 0 {
		ce = ce.Appendf("targetPrefix", "field must be populated")
	}
	if len(params.OriginPrefix) == 0 {
		ce = ce.Appendf("originPrefix", "field must be populated")
	}
	if len(params.LabelsValueName) == 0 {
		ce = ce.Appendf("labelsValueName", "field must be populated")
	}
	if len(params.PodIpValueName) == 0 {
		ce = ce.Appendf("podIpValueName", "field must be populated")
	}
	if len(params.PodNameValueName) == 0 {
		ce = ce.Appendf("podNameValueName", "field must be populated")
	}
	if len(params.HostIpValueName) == 0 {
		ce = ce.Appendf("hostIpValueName", "field must be populated")
	}
	if len(params.NamespaceValueName) == 0 {
		ce = ce.Appendf("namespaceValueName", "field must be populated")
	}
	if len(params.ServiceAccountValueName) == 0 {
		ce = ce.Appendf("serviceAccountValueName", "field must be populated")
	}
	if len(params.ServiceValueName) == 0 {
		ce = ce.Appendf("serviceValueName", "field must be populated")
	}
	if len(params.PodLabelForService) == 0 {
		ce = ce.Appendf("podLabelName", "field must be populated")
	}
	if len(params.TargetServiceInputName) == 0 {
		ce = ce.Appendf("targetServiceInputName", "field must be populated")
	}
	if len(params.ClusterDomainName) == 0 {
		ce = ce.Appendf("clusterDomainName", "field must be populated")
	} else if len(strings.Split(params.ClusterDomainName, ".")) != 3 {
		ce = ce.Appendf("clusterDomainName", "must have three segments, separated by '.' ('svc.cluster.local', for example)")
	}
	return
}

func (b *builder) BuildAttributesGenerator(env adapter.Env, c adapter.Config) (adapter.AttributesGenerator, error) {
	var clientErr error
	paramsProto := c.(*config.Params)
	b.once.Do(func() {
		refresh := paramsProto.CacheRefreshDuration
		controller, err := b.newCacheControllerFn(paramsProto.KubeconfigPath, refresh, env)
		if err != nil {
			clientErr = err
			return
		}
		b.pods = controller
		env.ScheduleDaemon(func() { b.pods.Run(b.stopChan) })
		// ensure that any request is only handled after
		// a sync has occurred
		cache.WaitForCacheSync(b.stopChan, b.pods.HasSynced)
	})
	if clientErr != nil {
		return nil, clientErr
	}
	kg := &kubegen{
		log:    env.Logger(),
		pods:   b.pods,
		params: *paramsProto,
	}
	return kg, nil
}

func newCacheFromConfig(kubeconfigPath string, refreshDuration time.Duration, env adapter.Env) (cacheController, error) {
	env.Logger().Infof("getting kubeconfig from: %#v", kubeconfigPath)
	config, err := getRESTConfig(kubeconfigPath)
	if err != nil || config == nil {
		return nil, fmt.Errorf("could not retrieve kubeconfig: %v", err)
	}
	env.Logger().Infof("getting k8s client from config")
	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create clientset for k8s: %v", err)
	}
	env.Logger().Infof("building new cache controller")
	return newCacheController(clientset, "", refreshDuration, env), nil
}

func getRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func (k *kubegen) Close() error { return nil }

func (k *kubegen) Generate(inputs map[string]interface{}) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	if uid, found := inputs[k.params.SourceUidInputName]; found {
		uidstr := uid.(string)
		if len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.SourcePrefix)
		}
	}
	if uid, found := inputs[k.params.TargetUidInputName]; found {
		uidstr := uid.(string)
		if len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.TargetPrefix)
		}
	}
	if uid, found := inputs[k.params.OriginUidInputName]; found {
		uidstr := uid.(string)
		if len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.OriginPrefix)
		}
	}
	if targetSvc, found := inputs[k.params.TargetServiceInputName]; found {
		svc := targetSvc.(string)
		if len(svc) > 0 {
			n, err := canonicalName(svc, defaultNamespace, k.params.ClusterDomainName)
			if err != nil {
				k.log.Warningf("could not canonicalize target service: %v", err)
			} else {
				values[valueName(k.params.TargetPrefix, k.params.ServiceValueName)] = n
			}
		}
	}
	return values, nil
}

func (k *kubegen) addValues(vals map[string]interface{}, uid, valPrefix string) {
	podKey := keyFromUID(uid)
	pod, err := k.pods.GetPod(podKey)
	if err != nil {
		k.log.Warningf("error getting pod for (uid: %s, key: %s): %v", uid, podKey, err)
	}
	addPodValues(vals, valPrefix, k.params, pod)
}

func keyFromUID(uid string) string {
	fullname := strings.TrimPrefix(uid, kubePrefix)
	if strings.Contains(fullname, ".") {
		parts := strings.Split(fullname, ".")
		if len(parts) == 2 {
			return fmt.Sprintf("%s/%s", parts[1], parts[0])
		}
	}
	return fullname
}

func addPodValues(m map[string]interface{}, prefix string, params config.Params, p *v1.Pod) {
	if p == nil {
		return
	}
	if len(p.Labels) > 0 {
		m[valueName(prefix, params.LabelsValueName)] = p.Labels
	}
	if len(p.Name) > 0 {
		m[valueName(prefix, params.PodNameValueName)] = p.Name
	}
	if len(p.Namespace) > 0 {
		m[valueName(prefix, params.NamespaceValueName)] = p.Namespace
	}
	if len(p.Spec.ServiceAccountName) > 0 {
		m[valueName(prefix, params.ServiceAccountValueName)] = p.Spec.ServiceAccountName
	}
	if len(p.Status.PodIP) > 0 {
		m[valueName(prefix, params.PodIpValueName)] = p.Status.PodIP
	}
	if len(p.Status.HostIP) > 0 {
		m[valueName(prefix, params.HostIpValueName)] = p.Status.HostIP
	}
	if app, found := p.Labels[params.PodLabelForService]; found {
		n, err := canonicalName(app, p.Namespace, params.ClusterDomainName)
		if err == nil {
			m[valueName(prefix, params.ServiceValueName)] = n
		}
	}
}

func valueName(prefix, value string) string {
	return fmt.Sprintf("%s%s", prefix, value)
}

// name format examples that can be currently canonicalized:
//
// "hello:80",
// "hello",
// "hello.default:80",
// "hello.default",
// "hello.default.svc:80",
// "hello.default.svc",
// "hello.default.svc.cluster:80",
// "hello.default.svc.cluster",
// "hello.default.svc.cluster.local:80",
// "hello.default.svc.cluster.local",
func canonicalName(service, namespace, clusterDomain string) (string, error) {
	if len(service) == 0 {
		return "", errors.New("invalid service name: cannot be empty")
	}
	// remove any port suffixes (ex: ":80")
	splits := strings.SplitN(service, ":", 2)
	s := splits[0]
	if len(s) == 0 {
		return "", fmt.Errorf("invalid service name '%s': starts with ':'", service)
	}
	// error on ip addresses for now
	if ip := net.ParseIP(s); ip != nil {
		return "", errors.New("invalid service name: cannot canonicalize ip addresses at this time")
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) == 1 {
		return fmt.Sprintf("%s.%s.%s", parts[0], namespace, clusterDomain), nil
	}
	if len(parts) == 2 {
		return fmt.Sprintf("%s.%s", s, clusterDomain), nil
	}

	domParts := strings.Split(clusterDomain, ".")
	nameParts := strings.Split(parts[2], ".")

	if len(nameParts) >= len(domParts) {
		return s, nil
	}
	for i := len(nameParts); i < len(domParts); i++ {
		s = fmt.Sprintf("%s.%s", s, domParts[i])
	}
	return s, nil
}
