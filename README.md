# gcp-ilb-redirect-controller

Automates the [creation of HTTP to HTTPS load balancer resources](https://cloud.google.com/load-balancing/docs/l7-internal/setting-up-http-to-https-redirect#partial-http-lb) for Internal HTTP(S) Load Balancers.

Example kubernetes manifests can be seen in [hack/app/](hack/app/).

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

