package s3

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Provider struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Provider(cfg aws.Config, bucket, region, id, secret string) S3Provider {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(id, secret, ""),
		)
	})

	return S3Provider{
		client: client,
		bucket: bucket,
		region: region,
	}
}

func (s *S3Provider) UploadImage(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := fmt.Sprintf("%s.jpg", uuid.NewString())
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	finalUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, filename)
	return finalUrl, nil
}
