package scenarios

// GatekeeperInstallationCommands returns the commands to install Gatekeeper using the official Helm chart
func GatekeeperInstallationCommands(releaseName string) [][]string {
	return [][]string{
		// Add the official Gatekeeper Helm repository
		{"helm", "repo", "add", "gatekeeper", "https://open-policy-agent.github.io/gatekeeper/charts"},
		// Update Helm repositories
		{"helm", "repo", "update"},
		// Install Gatekeeper using Helm
		{"helm", "install", releaseName, "gatekeeper/gatekeeper", "--namespace", "gatekeeper-system", "--create-namespace", "--wait", "--timeout=300s"},
	}
}

// GatekeeperUninstallCommands returns the commands to uninstall Gatekeeper
func GatekeeperUninstallCommands(releaseName string) [][]string {
	return [][]string{
		// Uninstall Gatekeeper Helm release
		{"helm", "uninstall", releaseName, "--namespace", "gatekeeper-system"},
		// Delete the namespace
		{"kubectl", "delete", "namespace", "gatekeeper-system", "--ignore-not-found"},
		// Clean up CRDs (as recommended in Gatekeeper docs)
		{"kubectl", "delete", "crd", "-l", "gatekeeper.sh/system=yes", "--ignore-not-found"},
	}
}

// HostPortConstraintTemplateScenario returns a ConstraintTemplate that forbids hostPort
func HostPortConstraintTemplateScenario() string {
	return `apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: k8xhostportforbidden
spec:
  crd:
    spec:
      names:
        kind: K8xHostPortForbidden
      validation:
        openAPIV3Schema:
          type: object
          properties:
            message:
              type: string
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8xhostportforbidden

        violation[{"msg": msg}] {
          input.review.object.kind == "Deployment"
          container := input.review.object.spec.template.spec.containers[_]
          port := container.ports[_]
          port.hostPort
          msg := sprintf("hostPort %v is forbidden by policy", [port.hostPort])
        }

        violation[{"msg": msg}] {
          input.review.object.kind == "Pod"
          container := input.review.object.spec.containers[_]
          port := container.ports[_]
          port.hostPort
          msg := sprintf("hostPort %v is forbidden by policy", [port.hostPort])
        }
`
}

// HostPortConstraintScenario returns a constraint that uses the HostPortForbidden template
func HostPortConstraintScenario() string {
	return `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8xHostPortForbidden
metadata:
  name: hostport-not-allowed
spec:
  match:
    kinds:
      - apiGroups: ["apps"]
        kinds: ["Deployment"]
      - apiGroups: [""]
        kinds: ["Pod"]
  parameters:
    message: "hostPort is not allowed in this cluster for security reasons"
`
}

// ForbiddenDeploymentScenario returns a deployment that violates the hostPort constraint
func ForbiddenDeploymentScenario() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: forbidden-hostport-app
  namespace: default
  labels:
    app: forbidden-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: forbidden-app
  template:
    metadata:
      labels:
        app: forbidden-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.20
        ports:
        - containerPort: 80
          hostPort: 8080  # This should be forbidden by Gatekeeper
        resources:
          limits:
            memory: "128Mi"
            cpu: "100m"
          requests:
            memory: "64Mi"
            cpu: "50m"
`
}
