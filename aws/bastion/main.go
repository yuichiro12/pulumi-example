package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		sg, err := ec2.NewSecurityGroup(ctx, "bastion", &ec2.SecurityGroupArgs{
			Name: pulumi.String("bastion"),
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		})
		if err != nil {
			return err
		}

		kp, err := ec2.NewKeyPair(ctx, "bastion", &ec2.KeyPairArgs{
			KeyName:   pulumi.String("ssh-keypair"),
			PublicKey: pulumi.String(config.Require(ctx, "sshPubKey")),
		})
		if err != nil {
			return err
		}

		ins, err := ec2.NewInstance(ctx, "bastion", &ec2.InstanceArgs{
			// amazon/al2022-ami-2022.0.20220531.0-kernel-5.15-arm64
			Ami:                      pulumi.String("ami-00ecc09e40aa9194b"),
			AssociatePublicIpAddress: pulumi.Bool(true),
			// t4g instances are based on arm64
			InstanceType:        ec2.InstanceType_T4g_Nano,
			VpcSecurityGroupIds: pulumi.StringArray{sg.ID()},
			KeyName:             kp.KeyName,
		})
		if err != nil {
			return err
		}

		eip, err := ec2.NewEip(ctx, "bastion", &ec2.EipArgs{
			Instance: ins.ID(),
		})
		if err != nil {
			return err
		}

		_, err = route53.NewRecord(ctx, "bastion", &route53.RecordArgs{
			Name:    pulumi.String("bastion.movapp.ch"),
			Type:    pulumi.String("A"),
			Ttl:     pulumi.Int(30),
			Records: pulumi.StringArray{eip.PublicIp},
			ZoneId:  pulumi.String(config.Require(ctx, "hostedZoneId")),
		})

		// https://qiita.com/yangci/items/ef2ab9b6f0d28bad0881#ec2-user%E3%81%AE%E5%89%8A%E9%99%A4

		return nil
	})
}
