# lokus

Export .local ingress hostnames via mDNS

Designed to be used with https://k3d.io/v5.6.0/ or other local kuberneteses that offer routable loadbalancers from the host.

## Install

Homebrew (macos/linux):

```bash
brew install mkmik/lokus/lokus
```

Anywhere else, from sources:

```bash
go install mkm.pub/lokus@latest
```

## Demo

Scenario: you have a few Ingress instances, possibly with different hostnames, all in the `.local` domain.

```console
$ kubectl -n mything get ing      
NAME               CLASS     HOSTS                     ADDRESS         PORTS     AGE
keycloak           traefik   keycloak.mything.local   192.168.228.2   80        5h42m
grpcingress        traefik   mything.local            192.168.228.2   80, 443   8d
metrics            traefik   mything.local            192.168.228.2   80, 443   8d
```

You run `lokus` somewhere in a shell that has access to your k8s cluster credentials

```console
$ lokus
2023/09/13 07:26:48 Serving ["keycloak.influxdb.local" "influxdb.local"] -> 192.168.228.2 using `dns-sd`
```

Now you can access your services as if they were local, using stable hostnames.
Stable hostnames are useful if you have scripts or config files that reference the hostnames.
This way you and your team mates can all have the same stable hostnames.

## Notes

- Tested only on macos (for now)
- When using Tailscale with split DNS, we can't use pure mDNS but we have to call into macos APIs. I tried 8 different libraries in Go and 5 in rust and found none that worked correctly with the Tailscale split DNS issue, so I resorted to just spawning the `dns-sd` subprocess (`dns-sd` is provided by macos).
