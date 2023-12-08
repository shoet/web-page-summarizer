package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var projectTag = "web-page-summarizer"

func CreateVPC(ctx *pulumi.Context, cidr string, resourceName string) (*ec2.Vpc, error) {
	return ec2.NewVpc(ctx, resourceName, &ec2.VpcArgs{
		CidrBlock:          pulumi.String(cidr),
		EnableDnsSupport:   pulumi.Bool(true),
		EnableDnsHostnames: pulumi.Bool(true),
		Tags:               createNameTag(resourceName),
	})
}

func CreateSubnet(
	ctx *pulumi.Context,
	vpc *ec2.Vpc,
	cidr string,
	// availabilityZone string,
	resourceName string,
) (*ec2.Subnet, error) {
	return ec2.NewSubnet(ctx, resourceName, &ec2.SubnetArgs{
		VpcId:     vpc.ID(),
		CidrBlock: pulumi.String(cidr),
		// AvailabilityZone: pulumi.String(availabilityZone),
		Tags: createNameTag(resourceName),
	})
}

func CreateIGW(
	ctx *pulumi.Context, vpc *ec2.Vpc, resourceName string,
) (*ec2.InternetGateway, error) {

	return ec2.NewInternetGateway(ctx, resourceName, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags:  createNameTag(resourceName),
	})
}

func CreatePublicRouteTable(
	ctx *pulumi.Context, vpc *ec2.Vpc, igw *ec2.InternetGateway, resourceName string,
) (*ec2.RouteTable, error) {
	return ec2.NewRouteTable(
		ctx, resourceName, &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: igw.ID(),
				},
			},
			Tags: createNameTag(resourceName),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc, igw}),
	)
}

func CreateRouteTableAssociation(
	ctx *pulumi.Context, routeTable *ec2.RouteTable, subnet *ec2.Subnet, resourceName string,
) (*ec2.RouteTableAssociation, error) {
	return ec2.NewRouteTableAssociation(
		ctx,
		resourceName,
		&ec2.RouteTableAssociationArgs{
			RouteTableId: routeTable.ID(),
			SubnetId:     subnet.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{routeTable, subnet}),
	)
}

func CreateSecurityGroupForECSTask(
	ctx *pulumi.Context, vpc *ec2.Vpc, resourceName string,
) (*ec2.SecurityGroup, error) {
	return ec2.NewSecurityGroup(
		ctx,
		resourceName,
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Description: pulumi.String("All outbound traffic"),
					Protocol:    pulumi.String("-1"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Tags: createNameTag(resourceName),
		})
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// AccountID ///////////////////////////////////////////////////////////////////
		caller, err := aws.GetCallerIdentity(ctx, nil, nil)
		if err != nil {
			return err
		}
		accountId := caller.AccountId

		// Region ///////////////////////////////////////////////////////////////////////
		region, err := aws.GetRegion(ctx, nil, nil)
		if err != nil {
			return err
		}

		// VPC //////////////////////////////////////////////////////////////////////////
		resourceName := fmt.Sprintf("%s-vpc", projectTag)
		vpc, err := CreateVPC(ctx, "10.3.0.0/16", resourceName)
		if err != nil {
			return fmt.Errorf("failed create vpc: %v", err)
		}
		ctx.Export(resourceName, vpc.ID())

		// Subnet /////////////////////////////////////////////////////////////////////////
		resourceName = fmt.Sprintf("%s-subnet-app-container", projectTag)
		subnetAppContainer, err := CreateSubnet(ctx, vpc, "10.3.0.0/24", resourceName)
		if err != nil {
			return fmt.Errorf("failed create subnet for App Container: %v", err)
		}
		ctx.Export(resourceName, subnetAppContainer.ID())

		// InternetGateway //////////////////////////////////////////////////////////
		resourceName = fmt.Sprintf("%s-igw", projectTag)
		igw, err := CreateIGW(ctx, vpc, resourceName)
		if err != nil {
			return fmt.Errorf("failed create igw: %v", err)
		}
		ctx.Export(resourceName, igw.ID())

		// ルートテーブル /////////////////////////////////////////////////////
		resourceName = fmt.Sprintf("%s-route-table-public", projectTag)
		publicRouteTable, err := CreatePublicRouteTable(ctx, vpc, igw, resourceName)
		if err != nil {
			return fmt.Errorf("failed create public route table: %v", err)
		}
		ctx.Export(resourceName, publicRouteTable.ID())

		// ルートテーブル 関連付け///////////////////////////////////////////////////////////
		resourceName = fmt.Sprintf("%s-route-table-association-app-container", projectTag)
		routeTableAssociationAppContainer, err := CreateRouteTableAssociation(
			ctx, publicRouteTable, subnetAppContainer, resourceName)
		if err != nil {
			return fmt.Errorf("failed create public route association for AppContainer: %v", err)
		}
		ctx.Export(resourceName, routeTableAssociationAppContainer.ID())

		// IAM //////////////////////////////////////////////////////////////
		// ECSタスクロール
		resourceName = fmt.Sprintf("%s-iam-role-for-ecs-task", projectTag)
		ecsTaskRole, err := iam.NewRole(
			ctx,
			resourceName,
			&iam.RoleArgs{
				AssumeRolePolicy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {
							"Service": "ecs-tasks.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}]
				}`),
				InlinePolicies: iam.RoleInlinePolicyArray{
					&iam.RoleInlinePolicyArgs{
						Name: pulumi.String("ecs-task-policy"),
						Policy: pulumi.String(`{
							"Version": "2012-10-17",
							"Statement": [
								{
								   "Effect": "Allow",
								   "Action": [
										"ssmmessages:CreateControlChannel",
										"ssmmessages:CreateDataChannel",
										"ssmmessages:OpenControlChannel",
										"ssmmessages:OpenDataChannel"
								   ],
								  "Resource": "*"
								},
								{
									"Effect": "Allow",
									"Action": [
										"dynamodb:GetItem",
										"dynamodb:Query",
										"dynamodb:Scan",
										"dynamodb:DeleteItem",
										"dynamodb:UpdateItem",
										"dynamodb:PutItem",
										"dynamodb:BatchWriteItem"
									],
								  "Resource": "*"
								},
								{
									"Effect": "Allow",
									"Action": [
										"sqs:SendMessage",
										"sqs:ReceiveMessage",
										"sqs:DeleteMessage",
										"sqs:GetQueueUrl"
									],
								  "Resource": "*"
								}
							]
						}`),
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed create iam role for ecs task: %v", err)
		}
		ctx.Export(resourceName, ecsTaskRole.ID())

		// ECSタスク実行用ロール
		resourceName = fmt.Sprintf("%s-iam-role-for-ecs-task-execute", projectTag)
		ecsTaskExecutionRole, err := iam.NewRole(
			ctx,
			resourceName,
			&iam.RoleArgs{
				AssumeRolePolicy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {
							"Service": "ecs-tasks.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}]
				}`),
				ManagedPolicyArns: pulumi.StringArray{
					pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
					pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
				},
				InlinePolicies: iam.RoleInlinePolicyArray{
					&iam.RoleInlinePolicyArgs{
						Name: pulumi.String("ecs-task-policy-logs"),
						Policy: pulumi.String(`{
							"Version": "2012-10-17",
							"Statement": [
								{
								   "Effect": "Allow",
								   "Action": [
										"logs:CreateLogGroup"
								   ],
								  "Resource": "*"
								},
							    {
								  "Effect": "Allow",
								  "Action": [
									"ssm:GetParameters",
									"secretsmanager:GetSecretValue",
									"kms:Decrypt"
								  ],
								  "Resource": "*"
								}
							]
						}`),
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed create iam role for ecs task execution: %v", err)
		}
		ctx.Export(resourceName, ecsTaskExecutionRole.ID())

		// セキュリティグループ securitygroup ///////////////////////////////////////////////
		resourceName = fmt.Sprintf("%s-sg-app-container", projectTag)
		securityGroupForECSTask, err := CreateSecurityGroupForECSTask(
			ctx, vpc, resourceName)
		if err != nil {
			return fmt.Errorf("failed create security group for ecs task: %v", err)
		}
		ctx.Export(resourceName, securityGroupForECSTask.ID())

		// ECS ////////////////////////////////////////////////////////////////////////
		// TaskDefinition
		taskDefinition, err := loadEcsContainerDefinition(
			"./container_definition.json", accountId, region.Name)
		if err != nil {
			return fmt.Errorf("failed load ecs task definition: %v", err)
		}
		resourceName = fmt.Sprintf("%s-ecs-task-definition", projectTag)
		ecsTaskDefinition, err := ecs.NewTaskDefinition(
			ctx,
			resourceName,
			&ecs.TaskDefinitionArgs{
				Family:                  pulumi.String("web-page-summarizer"),
				NetworkMode:             pulumi.String("awsvpc"),
				Cpu:                     pulumi.String("1024"),
				Memory:                  pulumi.String("3072"),
				TaskRoleArn:             ecsTaskRole.Arn,
				ExecutionRoleArn:        ecsTaskExecutionRole.Arn,
				RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
				ContainerDefinitions:    pulumi.String(taskDefinition),
			})
		ctx.Export(resourceName, ecsTaskDefinition.ID())

		return nil
	})
}

func createNameTag(tag string) pulumi.StringMap {
	return pulumi.StringMap{
		"Name": pulumi.String(tag),
	}
}

func loadFileToString(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed open file: %v", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed read file: %v", err)
	}
	return string(b), nil
}

func loadEcsContainerDefinition(
	path string, awsAccountId string, region string) (string, error) {
	type Values struct {
		AwsAccountId     string
		Region           string
		SecretsManagerId string
	}
	definition, err := loadFileToString(path)
	if err != nil {
		return "", fmt.Errorf("failed load ecs container definition: %v", err)
	}
	tmpl, err := template.New("ecsTaskDefinition").Parse(definition)
	if err != nil {
		return "", fmt.Errorf("failed parse ecs container definition: %v", err)
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, Values{
		AwsAccountId: awsAccountId,
		Region:       region,
	})
	if err != nil {
		return "", fmt.Errorf("failed execute ecs container definition: %v", err)
	}
	return buffer.String(), nil
}
