package resources

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func Test_isAWSLoadBalancerResource(t *testing.T) {
	type args struct {
		svc *v1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"nil as service",
			args{
				svc: nil,
			},
			false,
		},
		{
			"service is not a load balancer type",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeNodePort,
					},
				},
			},
			false,
		},
		{
			"service haven't finished the load balancer deployment (ingress is nil)",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: nil,
						},
					},
				},
			},
			false,
		},
		{
			"service haven't finished the load balancer deployment (ingress is empty)",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{},
						},
					},
				},
			},
			false,
		},
		{
			"service load balancer with empty hostname",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "",
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"service load balancer with non aws hostname",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "hello.test.com",
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"service load balancer with aws domain in the hostname",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "amazonaws.com.test.com",
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"service load balancer with aws domain",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "918238-99898.eu-west-1.elb.amazonaws.com",
								},
							},
						},
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAWSLoadBalancerResource(tt.args.svc); got != tt.want {
				t.Errorf("isAWSLoadBalancerResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
