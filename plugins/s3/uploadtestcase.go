package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type uploadTestcase struct {
	svc        *s3.S3
	sess       client.ConfigProvider
	bucketname string
}

func newUploadTestCase(svc *s3.S3, sess client.ConfigProvider) *uploadTestcase {
	return &uploadTestcase{
		svc:  svc,
		sess: sess,
	}
}

func (t *uploadTestcase) SetupOnce() {
	// for each test create a bucket to work with
	bucketname := strconv.Itoa(int(time.Now().UnixNano()))
	_, err := t.ensureBucket(bucketname)
	if err != nil {
		fmt.Println("err ensureBucket:", err.Error())
	}
	t.bucketname = bucketname
}

func (t *uploadTestcase) Setup() {
	log.Println("mock test setup")
}

func (t *uploadTestcase) Test() error {
	return t.upload()
}

func (t *uploadTestcase) Cleanup() {
	fmt.Println("cleaning up...")
	iter := s3manager.NewDeleteListIterator(t.svc, &s3.ListObjectsInput{
		Bucket: aws.String(t.bucketname),
	})

	if err := s3manager.NewBatchDeleteWithClient(t.svc).Delete(aws.BackgroundContext(), iter); err != nil {
		log.Println("Unable to delete objects from bucket:", t.bucketname, err)
	}
	_, err := t.svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(t.bucketname),
	})
	if err != nil {
		fmt.Println("Unable to delete bucket:", t.bucketname, err)
	}
	err = t.svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(t.bucketname),
	})
	if err != nil {
		fmt.Println("wait for delete bucket:", t.bucketname, err)
	}
}

func (t *uploadTestcase) ensureBucket(bucket string) (time.Duration, error) {
	start := time.Now()
	_, err := t.svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if !strings.Contains(err.Error(), "BucketAlreadyExists") {
			log.Println("create bucket err:", bucket, err)
			return 0, err
		}
		fmt.Println("bucket already exits... continuing")
		return 0, nil
	}

	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)
	err = t.svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Println("Error occurred while waiting for bucket to be created:", bucket)
		return 0, err
	}
	latency := time.Since(start)
	fmt.Printf("Bucket %q successfully created\n", bucket)
	return latency, nil
}

func (t *uploadTestcase) upload() error {
	fd, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Println("ioutil.Tempfile:", err)
	}
	defer fd.Close()

	_, err = fd.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Println("readall err", err)
	}
	info, err := fd.Stat()

	u, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		log.Println("rand uuid err", err)
	}

	upParams := &s3manager.UploadInput{
		Bucket: aws.String(t.bucketname),
		Key:    aws.String(fmt.Sprintf("%s/%s", string(u), info.Name())),
		Body:   bytes.NewReader(body),
	}
	uploader := s3manager.NewUploader(t.sess)
	_, err = uploader.Upload(upParams, func(u *s3manager.Uploader) {})

	if err != nil {
		if multierr, ok := err.(s3manager.MultiUploadFailure); ok {
			fmt.Println("Multi Error:", multierr.Code(), multierr.Message(), multierr.UploadID())
		} else {
			fmt.Println("Non Multi Error:", err.Error())
		}
		fmt.Println("failed to upload file:", err)
		return err
	}
	return nil
}
