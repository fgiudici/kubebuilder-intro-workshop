# kubebuilder-intro-workshop

## Prerquiesites

- golang
- access to Kubernetes cluster(Kind)

## Kubebuilder installation
```bash
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```
## Scaffolding the project
```bash
mkdir workshop/
cd workshop/
kubebuilder init --domain cattle.io --repo cattle.io/workshop
kubebuilder create api --group healthchecker --version v1 --kind HttpStatusPoller
```
Update controller tools in Makefile from:
```bash
CONTROLLER_TOOLS_VERSION ?= v0.12.0
```
to 
```bash
CONTROLLER_TOOLS_VERSION ?= v0.14.0
```

Run code generators
```
make generate
make manifests
```

## Example operator implementation

### API implementation
Add new fields to the API(api/v1/httpstatuspoller_types.go):

```golang
type HttpStatusPollerSpec struct {
	// +kubebuilder:validation:MinItems=1
	URLs []string `json:"urls"`

	// +optional
	IntervalSeconds int `json:"intervalSeconds,omitempty"`
}

// HttpStatusPollerStatus defines the observed state of HttpStatusPoller
type HttpStatusPollerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// StatusCodes tracks the HTTP return code of the checked URLs or -1 if
	// the URL can't be reached
	StatusCodes map[string]int `json:"statusCodes"`
}
```
  
Rerun code generators:
```bash
make generate && make manifests
```

### Controller implementation

Add a new function to poll URLs to controller(internal/controller/httpstatuspoller_controller.go):

```golang
func pollURLs(ctx context.Context, urls []string) map[string]int {
	log := log.FromContext(ctx)

	result := make(map[string]int)
	http.DefaultClient.Timeout = 2 * time.Second

	for _, u := range urls {
		resp, err := http.Get(u)
		if err != nil {
			if err.(*url.Error).Timeout() {
				log.Info("URL is unreachable", "url", u)
			} else {
				log.Error(err, "Failed to poll URL", "url", u)
			}
			result[u] = -1
			continue
		}
		defer resp.Body.Close()
		log.Info("URL is reachable", "url", u, "HTTP Status code", resp.Status)
		result[u] = resp.StatusCode
	}

	return result
}
```

Modify `func (r *HttpStatusPollerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)`:
```golang
func (r *HttpStatusPollerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	httpStatusPoller := &healthcheckerv1.HttpStatusPoller{}
	if err := r.Get(ctx, req.NamespacedName, httpStatusPoller); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patch := client.MergeFrom(httpStatusPoller.DeepCopy())

	// Poll the URLs specified in the CR
	result := pollURLs(ctx, httpStatusPoller.Spec.URLs)

	// Update the status of the CR with the results of the polling
	httpStatusPoller.Status.StatusCodes = result

	if err := r.Status().Patch(ctx, httpStatusPoller, patch); err != nil {
		return ctrl.Result{}, err
	}

	pollingInterval := time.Duration(httpStatusPoller.Spec.IntervalSeconds) * time.Second
	if httpStatusPoller.Spec.IntervalSeconds == 0 {
		pollingInterval = time.Minute
	}

	return ctrl.Result{RequeueAfter: pollingInterval}, nil
}
```

## Run the operator

```bash
kind create cluster
make install
make run
```

## Apply the example
```bash
kubectl apply -f - <<EOF
apiVersion: healthchecker.cattle.io/v1
kind: HttpStatusPoller
metadata:
  name: httpstatuspoller-sample
spec:
  urls:
  - https://google.com
  - https://1.2.3.4
  - https://doesnt.exist
EOF
```

