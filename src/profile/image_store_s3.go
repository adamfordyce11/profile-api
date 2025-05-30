package profile

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3ImageStore struct {
	Client     *s3.Client
	BucketName string
}

func (s *S3ImageStore) InitBucketAndCORS(ctx context.Context) error {
	// Check if bucket exists
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.BucketName),
	})
	if err != nil {
		endpoint := os.Getenv("AWS_S3_ENDPOINT")
		if endpoint != "" {
			// LocalStack: do NOT set CreateBucketConfiguration
			_, createErr := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(s.BucketName),
			})
			if createErr != nil {
				return fmt.Errorf("unable to create S3 bucket: %w", createErr)
			}
		} else {
			region := os.Getenv("AWS_REGION")
			input := &s3.CreateBucketInput{
				Bucket: aws.String(s.BucketName),
			}
			if region != "" && region != "us-east-1" {
				input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
					LocationConstraint: s3types.BucketLocationConstraint(region),
				}
			}
			_, createErr := s.Client.CreateBucket(ctx, input)
			if createErr != nil {
				return fmt.Errorf("unable to create S3 bucket: %w", createErr)
			}
		}
	}

	// Set CORS policy
	_, corsErr := s.Client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(s.BucketName),
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: []s3types.CORSRule{
				{
					AllowedHeaders: []string{"*"},
					AllowedMethods: []string{"GET"},
					AllowedOrigins: []string{"*"},
					ExposeHeaders:  []string{},
				},
			},
		},
	})
	if corsErr != nil {
		return fmt.Errorf("unable to set CORS on S3 bucket: %w", corsErr)
	}

	return nil
}

func (s *S3ImageStore) SaveImage(userID, filename string, file multipart.File) (string, error) {
	imageName := fmt.Sprintf("%s-%s", userID, filename)

	// Upload the file to S3
	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(imageName),
		Body:   file,
		ACL:    "public-read", // For LocalStack and public S3 access
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %w", err)
	}

	// Construct the public URL
	// For AWS S3: https://{bucket}.s3.{region}.amazonaws.com/{key}
	// For LocalStack: http://localhost:4566/{bucket}/{key}
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	var imageURL string
	if endpoint != "" {
		// Replace "localstack" with "localhost" for URLs returned to the frontend
		publicEndpoint := endpoint
		if strings.Contains(endpoint, "localstack") {
			publicEndpoint = strings.Replace(endpoint, "localstack", "localhost", 1)
		}
		imageURL = fmt.Sprintf("%s/%s/%s", publicEndpoint, s.BucketName, imageName)
	} else {
		region := os.Getenv("AWS_REGION")
		imageURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.BucketName, region, imageName)
	}

	return imageURL, nil
}