package provider

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func EC2Tags(instanceID string) map[string]string {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := ec2.New(sess)
	input := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(instanceID),
				},
			},
		},
	}

	result, err := svc.DescribeTags(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	tags := make(map[string]string)
	for _, tag := range result.Tags {
		tags[*tag.Key] = *tag.Value
	}
	return tags
}
