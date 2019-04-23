package providerclient

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func Test_getAWSLoadBalancerName(t *testing.T) {
	type args struct {
		svc *v1.Service
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test with one dash in the name",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "aa59f0ca83-7455.eu-west-1.elb.amazonaws.com",
								},
							},
						},
					},
				},
			},
			"aa59f0ca83",
		},
		{
			"test with two dash in the name",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "aa59f0-ca83-7455.eu-west-1.elb.amazonaws.com",
								},
							},
						},
					},
				},
			},
			"aa59f0-ca83",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAWSLoadBalancerName(tt.args.svc); got != tt.want {
				t.Errorf("getAWSLoadBalancerName() = %v, want %v", got, tt.want)
			}
		})
	}
}
