# prometheus-proxy

A reverse proxy for prometheus targets to authenticate metric requests.

## The problem

When deploying prometheus metric collection across the internet, you may want to restrict access to target endpoints.

![prometheus-firewall](images/prometheus-firewall.png)

With a static instances, this can be accomplished with firewall rules. This solution does not work when source addresses are dynamic, and the prometheus collectors change.

## A solution

The prometheus-proxy allows hosts to expose a single port for all local exporters, and secure access through TLS certificates.

![prometheus-proxy](images/prometheus-proxy.png)

In the new configuration, prometheus connect to the `prometheus-proxy` with a client certificate signed by a central CA. Each `prometheus-proxy` instances runs with a server certificate signed byt the same CA.

Prometheus validates the server certificate is valid by referencing the CA's public certificate.

Prometheus-proxy validates the prometheus client certificate against the same CA.

# Advantages

- No more firewall rules (ok, 1 for the prometheus-proxy)
- Fully dynamic endpoints, we just need to update DNS and as long as the cerfiicates validate, everything continues to work
