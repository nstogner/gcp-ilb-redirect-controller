# gcp-ilb-redirect-controller

**State**: Experimental

Automates the [creation of HTTP to HTTPS load balancer resources](https://cloud.google.com/load-balancing/docs/l7-internal/setting-up-http-to-https-redirect#partial-http-lb) for Internal HTTP(S) Load Balancers.

Example kubernetes manifests can be seen in [hack/app/](hack/app/). Note the referenced TLS certificate (Ingress annotation `ingress.gcp.kubernetes.io/pre-shared-cert`) and internal static IP (`kubernetes.io/ingress.regional-static-ip-name`) must be created.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: foo
  namespace: default
  annotations:
    ingress.gcp.kubernetes.io/pre-shared-cert: "foo-cert"
    kubernetes.io/ingress.class: "gce-internal"
    kubernetes.io/ingress.allow-http: "false" # Needed for the gce-ingress controller to create a HTTPS load balancer.
    kubernetes.io/ingress.regional-static-ip-name: "foo-1"
    networking.gke.io/ilb-https-redirect: "ok" # <-- This controller will look at this annotation and add a Forwarding Rule that redirects HTTP to HTTPS.
```

## Install in Cluster

Deploy controller.

```sh
kubectl apply -f deploy/
```

Authorize controller with a GKE Workload Identity.

```sh
export PROJECT=<your-gcp-project>

export SA_NAME=k8s-ilb-redirect-controller
export KSA_NAME=ilb-redirect-controller-controller-manager
export NAMESPACE=ilb-redirect-controller-system
```

```sh
gcloud iam service-accounts create $SA_NAME --project=$PROJECT

gcloud projects add-iam-policy-binding $PROJECT \
    --member "serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
    --role "roles/compute.networkAdmin"

gcloud iam service-accounts add-iam-policy-binding ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"

kubectl annotate serviceaccount $KSA_NAME \
    --namespace $NAMESPACE \
    iam.gke.io/gcp-service-account=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com
```

## Run Locally

When you run locally (instead of installing in the cluster), the controller will use local Kubeconfig credentials to connect to Kubernetes and local GCP credentials to operate against GCP.

```sh
make run PROJECT=<your-gcp-project> REGION=<your-gke-region>
```

