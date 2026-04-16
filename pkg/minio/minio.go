package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"go-fx-template/internal/utils/text"
)

type Config struct {
	Endpoint              string
	AccessKeyID           string
	SecretAccessKey       string
	UseSSL                bool
	BucketName            string
	SetBucketPublicPolicy bool
}

const connectTimeout = 2 * time.Second

func NewClient(cfg Config) (*minio.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	fmt.Print("Connecting to MinIO S3...")

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to initialize client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to connect: %w", err)
	}
	if !exists {
		if err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			fmt.Println()
			return nil, fmt.Errorf("MinIO: failed to create bucket '%s': %w", cfg.BucketName, err)
		}
	}

	if cfg.SetBucketPublicPolicy {
		allowPublicReadPolicy := fmt.Sprintf(`{
		  "Version": "2012-10-17",
		  "Statement": [
			{
			  "Sid": "PublicReadObjects",
			  "Effect": "Allow",
			  "Principal": { "AWS": [ "*" ] },
			  "Action": [ "s3:GetObject" ],
			  "Resource": [ "arn:aws:s3:::%s/*" ]
			}
		  ]
		}`, cfg.BucketName)

		if err = client.SetBucketPolicy(ctx, cfg.BucketName, allowPublicReadPolicy); err != nil {
			fmt.Println()
			return nil, fmt.Errorf("MinIO: failed to set public read policy for bucket '%s': %w", cfg.BucketName, err)
		}
	}

	fmt.Println(text.Green(" Done."))
	return client, nil
}
