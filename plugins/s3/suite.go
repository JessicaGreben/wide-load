package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/jessicagreben/wide-load/pkg/loader"
)

type testsuite struct {
	testcases []loader.Testcase

	sess client.ConfigProvider
	svc  *s3.S3
}

func (suite *testsuite) Init() error {
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		return fmt.Errorf("env var AWS_ACCESS_KEY_ID is required")
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		return fmt.Errorf("env var AWS_SECRET_ACCESS_KEY is required")
	}
	awsS3Endpoint := os.Getenv("S3_ENDPOINT")
	if awsS3Endpoint == "" {
		return fmt.Errorf("env var S3_ENDPOINT is required")
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("anywhere"),
		Credentials: credentials.NewStaticCredentials(
			awsAccessKeyID,
			awsSecretAccessKey,
			"",
		),
		Endpoint: aws.String(awsS3Endpoint),
		// S3ForcePathStyle: IsPathStyle,
		// LogLevel:         aws.LogLevel(aws.LogDebug),
	})
	if err != nil {
		fmt.Println("err new session:", err)
		return err
	}

	suite.sess = sess
	suite.svc = s3.New(sess)
	return nil
}

func (suite *testsuite) AddTests() int {
	suite.testcases = append(suite.testcases, newUploadTestCase(suite.svc, suite.sess))
	return len(suite.testcases)
}

func (suite *testsuite) Tests() []loader.Testcase {
	return suite.testcases
}
