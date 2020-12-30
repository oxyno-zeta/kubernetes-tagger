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
			"test with one dash in the name and internal prefix",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "internal-aa59f0ca83-7455.eu-west-1.elb.amazonaws.com",
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
		{
			"test with two dash in the name and internal prefix",
			args{
				svc: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								v1.LoadBalancerIngress{
									Hostname: "internal-aa59f0-ca83-7455.eu-west-1.elb.amazonaws.com",
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

func Test_getVolumeIDFromPersistentVolume(t *testing.T) {
	type args struct {
		pv *v1.PersistentVolume
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"AWS EBS Spec with availability zone in",
			args{
				pv: &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
								VolumeID: "aws://eu-west-1a/vol-test12131213",
							},
						},
					},
				},
			},
			"vol-test12131213",
			false,
		},
		{
			"AWS EBS Spec without availability zone in",
			args{
				pv: &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
								VolumeID: "aws:///vol-test12131213",
							},
						},
					},
				},
			},
			"vol-test12131213",
			false,
		},
		{
			"AWS EBS Spec with a / at the end",
			args{
				pv: &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
								VolumeID: "aws:///vol-test12131213/",
							},
						},
					},
				},
			},
			"vol-test12131213",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVolumeIDFromPersistentVolume(tt.args.pv)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVolumeIDFromPersistentVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getVolumeIDFromPersistentVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
