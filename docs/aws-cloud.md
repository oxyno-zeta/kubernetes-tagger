# AWS Cloud

## Configuration

Kubernetes Tagger support AWS cloud. To enable it, just put the following key in the configuration file:

```yaml
provider: aws
```

Moreover, this cloud is actually the default one enabled in Kubernetes Tagger.

## IAM Policies

Here is the AMI Policies that Kubernetes Tagger needs in AWS:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:RemoveTags",
                "elasticloadbalancing:DescribeTags",
                "elasticloadbalancing:AddTags",
                "ec2:DescribeVolumes"
            ],
            "Resource": "*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:RemoveTags",
                "ec2:DeleteTags",
                "ec2:CreateTags",
                "elasticloadbalancing:AddTags"
            ],
            "Resource": [
                "arn:aws:ec2:*:*:volume/*",
                "arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*",
                "arn:aws:elasticloadbalancing:*:*:loadbalancer/net/*/*"
            ]
        }
    ]
}
```
