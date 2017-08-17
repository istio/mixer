// Copyright 2017 Istio Authors.
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
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // needed for auth
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"istio.io/mixer/adapter/kubernetes/config"
	"istio.io/mixer/pkg/adapter"
)

type (
	builder struct {
		adapter.DefaultBuilder
		sync.Mutex

		stopChan             chan struct{}
		pods                 cacheController
		newCacheControllerFn controllerFactoryFn
		needsCacheInit       bool
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
	kubePrefix = "kubernetes://"

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
	clusterDomain        = "svc.cluster.local"
	podServiceLabel      = "app"
	istioPodServiceLabel = "istio"

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
		ClusterDomainName:       clusterDomain,
		PodLabelForService:      podServiceLabel,
		PodLabelForIstioService: istioPodServiceLabel,
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
	return &builder{
		adapter.NewDefaultBuilder(name, desc, conf),
		sync.Mutex{},
		stopChan,
		nil,
		cacheFactory,
		true,
	}
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
		ce = ce.Appendf("podLabelForService", "field must be populated")
	}
	if len(params.PodLabelForIstioService) == 0 {
		ce = ce.Appendf("podLabelForIstioService", "field must be populated")
	}
	if len(params.ClusterDomainName) == 0 {
		ce = ce.Appendf("clusterDomainName", "field must be populated")
	} else if len(strings.Split(params.ClusterDomainName, ".")) != 3 {
		ce = ce.Appendf("clusterDomainName", "must have three segments, separated by '.' ('svc.cluster.local', for example)")
	}
	return
}

func (b *builder) BuildAttributesGenerator(env adapter.Env, c adapter.Config) (adapter.AttributesGenerator, error) {
	paramsProto := c.(*config.Params)
	b.Lock()
	defer b.Unlock()
	if b.needsCacheInit {
		refresh := paramsProto.CacheRefreshDuration
		path, exists := os.LookupEnv("KUBECONFIG")
		if !exists {
			path = paramsProto.KubeconfigPath
		}
		controller, err := b.newCacheControllerFn(path, refresh, env)
		if err != nil {
			return nil, err
		}
		b.pods = controller
		env.ScheduleDaemon(func() { b.pods.Run(b.stopChan) })
		// ensure that any request is only handled after
		// a sync has occurred
		env.Logger().Infof("Waiting for kubernetes cache sync...")
		if success := cache.WaitForCacheSync(b.stopChan, b.pods.HasSynced); !success {
			b.stopChan <- struct{}{}
			return nil, errors.New("cache sync failure")
		}
		env.Logger().Infof("Cache sync successful.")
		b.needsCacheInit = false
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
	return newCacheController(clientset, refreshDuration, env), nil
}

func getRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func (k *kubegen) Close() error { return nil }

func (k *kubegen) Generate(inputs map[string]interface{}) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	if uid, found := inputs[k.params.SourceUidInputName]; found {
		uidstr, ok := uid.(string)
		if ok && len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.SourcePrefix)
		}
	}
	if uid, found := inputs[k.params.TargetUidInputName]; found {
		uidstr, ok := uid.(string)
		if ok && len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.TargetPrefix)
		}
	}
	if uid, found := inputs[k.params.OriginUidInputName]; found {
		uidstr, ok := uid.(string)
		if ok && len(uidstr) > 0 {
			k.addValues(values, uidstr, k.params.OriginPrefix)
		}
	}
	return values, nil
}

func (k *kubegen) addValues(vals map[string]interface{}, uid, valPrefix string) {
	podKey := keyFromUID(uid)
	pod, found := k.pods.GetPod(podKey)
	if !found {
		k.log.Warningf("could not find pod for (uid: %s, key: %s)", uid, podKey)
		return
	}
	svc, found := k.pods.GetServiceForPod(key(pod.Namespace, pod.Name))
	if !found {
		k.log.Warningf("error finding service for pod '%s'", pod.Name)
	}
	addPodValues(vals, valPrefix, k.params, pod, svc)
}

func keyFromUID(uid string) string {
	if ip := net.ParseIP(uid); ip != nil {
		return uid
	}
	fullname := strings.TrimPrefix(uid, kubePrefix)
	if strings.Contains(fullname, ".") {
		parts := strings.Split(fullname, ".")
		if len(parts) == 2 {
			return key(parts[1], parts[0])
		}
	}
	return fullname
}

func addPodValues(m map[string]interface{}, prefix string, params config.Params, p *v1.Pod, svc *v1.Service) {
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
	if svc != nil {
		n, err := canonicalName(svc.Name, svc.Namespace, params.ClusterDomainName)
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
