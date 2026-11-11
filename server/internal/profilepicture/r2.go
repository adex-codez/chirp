package profilepicture

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3            *s3.Client
	bucket        string
	publicBaseURL url.URL
}

var objectIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func NewClient(cfg config.R2Config) *Client {
	client := s3.NewFromConfig(aws.Config{
		Region:      "auto",
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://" + cfg.AccountID + ".r2.cloudflarestorage.com")
		options.UsePathStyle = true
		options.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		options.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	return &Client{s3: client, bucket: cfg.Bucket, publicBaseURL: cfg.PublicBaseURL}
}

func (c *Client) Upload(ctx context.Context, userID string, picture JPEG) (string, error) {
	var objectID [16]byte
	if _, err := rand.Read(objectID[:]); err != nil {
		return "", err
	}
	objectID[6] = objectID[6]&0x0f | 0x40
	objectID[8] = objectID[8]&0x3f | 0x80
	name := fmt.Sprintf("%x-%x-%x-%x-%x.jpg", objectID[:4], objectID[4:6], objectID[6:8], objectID[8:10], objectID[10:])
	key := "profile-pictures/" + userID + "/" + name
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(picture.data),
		ContentLength: aws.Int64(int64(len(picture.data))),
		ContentType:   aws.String("image/jpeg"),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", err
	}
	publicURL := c.publicBaseURL
	publicURL.Path = "/" + key
	return publicURL.String(), nil
}

func (c *Client) DeleteOwned(ctx context.Context, userID, previousURL string) error {
	previous, err := url.Parse(previousURL)
	if err != nil || previous.Scheme != c.publicBaseURL.Scheme || previous.Host != c.publicBaseURL.Host || previous.User != nil || previous.Opaque != "" || previous.RawQuery != "" || previous.ForceQuery || previous.Fragment != "" || strings.ContainsAny(previousURL, "%?#") {
		return nil
	}
	name, owned := strings.CutPrefix(previous.Path, "/profile-pictures/"+userID+"/")
	if !owned || !strings.HasSuffix(name, ".jpg") || !objectIDPattern.MatchString(strings.TrimSuffix(name, ".jpg")) {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err = c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(strings.TrimPrefix(previous.Path, "/")),
	})
	return err
}
