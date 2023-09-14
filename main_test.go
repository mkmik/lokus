package main

import (
	"fmt"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
)

func TestMultipleLoadBalancerIPs(t *testing.T) {

	fakeIngress := []networkingv1.Ingress{

		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fake",
				Namespace: "test",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Ingress",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					networkingv1.IngressRule{
						Host: "fake.local",
					}},
			},
			Status: networkingv1.IngressStatus{
				LoadBalancer: networkingv1.IngressLoadBalancerStatus{
					Ingress: []networkingv1.IngressLoadBalancerIngress{
						networkingv1.IngressLoadBalancerIngress{IP: "192.168.48.4"},
						networkingv1.IngressLoadBalancerIngress{IP: "192.168.48.5"},
					},
				},
			},
		},
	}

	names, ip, err := generateHosts(fakeIngress)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	fmt.Println(names, ip)
	expectedNames := []string{"fake.local"}
	expectedIp := "192.168.48.4" // The first address we find

	if !slices.Equal(names, expectedNames) {
		t.Fatalf("expected: %v, got: %v", expectedNames, names)
	}

	if ip != expectedIp {
		t.Fatalf("expected: %v, got: %v", expectedIp, ip)
	}

}
