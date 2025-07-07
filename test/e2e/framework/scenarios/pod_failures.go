package scenarios

import (
	"fmt"
	"time"
)

// CrashLoopBackoffScenario returns a manifest for a pod that will crash loop
func CrashLoopBackoffScenario() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: k8x-test
---
apiVersion: v1
kind: Pod
metadata:
  name: crash-loop-pod
  namespace: k8x-test
  labels:
    app: crash-loop
spec:
  containers:
  - name: crash-loop-container
    image: busybox:latest
    command: ["/bin/sh", "-c", "echo 'I am going to crash now'; sleep 5; exit 1"]
    resources:
      limits:
        memory: "64Mi"
        cpu: "100m"
  restartPolicy: Always
`
}

// MissingResourcesScenario returns a manifest for a pod that will be pending due to insufficient resources
func MissingResourcesScenario() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: k8x-test
---
apiVersion: v1
kind: Pod
metadata:
  name: resource-hungry-pod
  namespace: k8x-test
  labels:
    app: resource-hungry
spec:
  containers:
  - name: resource-hungry-container
    image: busybox:latest
    command: ["/bin/sh", "-c", "sleep 3600"]
    resources:
      requests:
        memory: "8Gi"
        cpu: "4"
      limits:
        memory: "8Gi"
        cpu: "4"
  restartPolicy: Always
`
}

// ImagePullBackoffScenario returns a manifest for a pod with an invalid image
func ImagePullBackoffScenario() string {
	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: k8x-test
---
apiVersion: v1
kind: Pod
metadata:
  name: image-pull-error-pod
  namespace: k8x-test
  labels:
    app: image-pull-error
spec:
  containers:
  - name: image-pull-error-container
    image: nonexistentimage%d:latest
    resources:
      limits:
        memory: "64Mi"
        cpu: "100m"
  restartPolicy: Always
`, time.Now().Unix())
}

// ConfigMapMissingScenario returns a manifest for a pod that depends on a missing ConfigMap
func ConfigMapMissingScenario() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: k8x-test
---
apiVersion: v1
kind: Pod
metadata:
  name: missing-config-pod
  namespace: k8x-test
  labels:
    app: missing-config
spec:
  containers:
  - name: missing-config-container
    image: busybox:latest
    command: ["/bin/sh", "-c", "sleep 3600"]
    volumeMounts:
    - name: config-volume
      mountPath: /config
    resources:
      limits:
        memory: "64Mi"
        cpu: "100m"
  volumes:
  - name: config-volume
    configMap:
      name: non-existent-config
  restartPolicy: Always
`
}
