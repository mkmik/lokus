package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
	"golang.org/x/sync/errgroup"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// set by goreleaser
var version = "(devel)"

const (
	debugServiceName = "debug-service"
	debugPortName    = "http"
)

type Context struct {
	*CLI
}

type CLI struct {
	Kubeconfig string `name:"kubeconfig"`
	Namespace  string `name:"namespace" short:"n" help:"specific namespace or else all namespaces if empty"`

	Version kong.VersionFlag `name:"version" help:"Print version information and quit"`
}

func kubeconfig(kubeconfigPath string) (string, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	switch {
	case kubeconfig != "":
		return kubeconfig, nil
	case kubeconfigPath != "":
		return kubeconfigPath, nil
	default:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".kube", "config"), nil
	}
}

func (cmd *CLI) Run(cli *Context) error {
	kubeconfigPath, err := kubeconfig(cmd.Kubeconfig)
	if err != nil {
		return err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Use the client to get the Ingresses in the current namespace
	ingresses, err := clientset.NetworkingV1().Ingresses(cmd.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	return advertise(ingresses.Items)
}

func deduplicate(data []string) []string {
	m := make(map[string]bool)
	for _, item := range data {
		m[item] = true
	}
	result := make([]string, 0, len(m))
	for item := range m {
		result = append(result, item)
	}
	return result
}

func advertise(ingresses []networkingv1.Ingress) error {
	var (
		ip    string
		names []string
	)
	for _, ingress := range ingresses {
		for _, rule := range ingress.Spec.Rules {
			if strings.HasSuffix(rule.Host, ".local") {
				names = append(names, rule.Host)
			}
		}
		for _, lb := range ingress.Status.LoadBalancer.Ingress {

			// Address has previously been set when there are multiple LB IPs,
			// this is okay to proceed.
			if ip == lb.IP {
				continue
			}

			if ip != lb.IP {
				if ip == "" {
					ip = lb.IP
				}

			} else {
				return fmt.Errorf("lokus works only if all matching ingresses use the same loadbalancer")
			}
		}
	}
	names = deduplicate(names)

	if runtime.GOOS == "darwin" {
		return advertiseMacHack(names, ip)
	}

	addr, err := net.ResolveUDPAddr("udp4", mdns.DefaultAddress)
	// addr, err := net.ResolveUDPAddr("udp6", "[ff02::0]:5353")

	if err != nil {
		return err
	}

	l, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}

	_, err = mdns.Server(ipv4.NewPacketConn(l), &mdns.Config{
		LocalNames:   names,
		LocalAddress: net.ParseIP(ip),
	})
	if err != nil {
		return err
	}
	log.Printf("Serving %q -> %s over mDNS listening on %s", names, ip, addr)
	select {}
}

func advertiseMacHack(names []string, ip string) error {
	g, _ := errgroup.WithContext(context.Background())

	for _, name := range names {
		name := name
		_ = name
		g.Go(func() error {
			cmd := exec.Command("dns-sd", "-P", name, "_http", "local", "80", name, ip)
			return cmd.Run()
		})
	}
	log.Printf("Serving %q -> %s using `dns-sd`", names, ip)
	log.Printf("(Running dns-sd instead of mDNS as a workaround for Tailscale split DNS issue)")
	if err := g.Wait(); err != nil {
		return err
	}

	log.Println("Should not end successfully")
	return nil
}

func getVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	// otherwise fallback to the version set by goreleaser
	return version
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Description(`For each Ingress resource in your local k3d cluster that uses a ".local" domain name, 
lokus creates a DNS record pointing to the internal load balancer IP. It uses unprivileged mDNS so you don't have to edit /etc/hosts`),
		kong.UsageOnError(),
		kong.Vars{
			"version": getVersion(),
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	err := ctx.Run(&Context{CLI: &cli})
	ctx.FatalIfErrorf(err)
}
