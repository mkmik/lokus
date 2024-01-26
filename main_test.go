package main

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
)

func TestGenerateHosts(t *testing.T) {
	type expectedHost struct {
		names []string
		ip    string
	}
	tests := []struct {
		name          string
		fakeIngresses []networkingv1.Ingress
		expectedHosts []expectedHost
	}{
		{
			name: "MultipleLoadBalancerIPs",
			fakeIngresses: []networkingv1.Ingress{
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
							{
								Host: "fake.local",
							},
						},
					},
					Status: networkingv1.IngressStatus{
						LoadBalancer: networkingv1.IngressLoadBalancerStatus{
							Ingress: []networkingv1.IngressLoadBalancerIngress{
								{IP: "192.168.48.4"},
								{IP: "192.168.48.5"},
							},
						},
					},
				},
			},
			expectedHosts: []expectedHost{
				{names: []string{"fake.local"}, ip: "192.168.48.4"},
			},
		},
		{
			name: "MultipleHosts",
			fakeIngresses: []networkingv1.Ingress{
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
							{
								Host: "one.local",
							},
							{
								Host: "two.local",
							},
						},
					},
					Status: networkingv1.IngressStatus{
						LoadBalancer: networkingv1.IngressLoadBalancerStatus{
							Ingress: []networkingv1.IngressLoadBalancerIngress{
								{IP: "192.168.48.4"},
							},
						},
					},
				},
			},
			expectedHosts: []expectedHost{
				{names: []string{"one.local", "two.local"}, ip: "192.168.48.4"},
			},
		},
		{
			name: "MultipleIngresses",
			fakeIngresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "one",
						Namespace: "test",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Ingress",
					},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "one.local",
							},
						},
					},
					Status: networkingv1.IngressStatus{
						LoadBalancer: networkingv1.IngressLoadBalancerStatus{
							Ingress: []networkingv1.IngressLoadBalancerIngress{
								{IP: "192.168.48.4"},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "two",
						Namespace: "test",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Ingress",
					},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "two.local",
							},
						},
					},
					Status: networkingv1.IngressStatus{
						LoadBalancer: networkingv1.IngressLoadBalancerStatus{
							Ingress: []networkingv1.IngressLoadBalancerIngress{
								{IP: "192.168.48.4"},
							},
						},
					},
				},
			},
			expectedHosts: []expectedHost{
				{names: []string{"one.local", "two.local"}, ip: "192.168.48.4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hosts, err := generateHosts(tt.fakeIngresses)
			if err != nil {
				t.Fatalf("expected no error, got %+v", err)
			}

			if got, want := len(hosts), len(tt.expectedHosts); got != want {
				t.Fatalf("got: %v, want: %v", got, want)
			}

			for i, host := range hosts {
				if got, want := host.names, tt.expectedHosts[i].names; !slices.Equal(got, want) {
					t.Fatalf("got: %v, want: %v", got, want)
				}

				if got, want := host.ip, tt.expectedHosts[i].ip; got != want {
					t.Fatalf("got: %v, want: %v", got, want)
				}
			}
		})
	}
}
