package main

import (
	"time"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Cli *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presigned_client := s3.NewPresignClient(s3Cli)
	getObjectInput := &s3.GetObjectInput{
		Bucket:			&bucket,
		Key:			&key,
	}
	presigned_req, err := presigned_client.PresignGetObject(context.Background(), getObjectInput, s3.WithPresignExpires(expireTime))
	if err != nil {return "", err}

	return presigned_req.URL, nil
}