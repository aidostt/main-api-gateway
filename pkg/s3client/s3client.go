package s3client

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"mime/multipart"
)

type S3Client struct {
	uploader *s3manager.Uploader
	bucket   string
}

func NewS3Client(region, httpAddress, accessKey, privateKey, bucket string) *S3Client {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, privateKey, ""),
		MaxRetries:  aws.Int(5),
		//Endpoint:    aws.String(httpAddress),
	}))
	uploader := s3manager.NewUploader(sess)

	return &S3Client{
		uploader: uploader,
		bucket:   bucket,
	}
}

func (s *S3Client) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}

	fileName := fileHeader.Filename
	result, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
		Body:   buf,
	})
	if err != nil {
		return "", err
	}
	return result.Location, nil
}
