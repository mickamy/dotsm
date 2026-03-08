package sm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// API is the subset of the Secrets Manager client used by this package.
type API interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	PutSecretValue(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
}

// Client wraps the Secrets Manager API for dotenv-compatible operations.
type Client struct {
	api API
}

// New creates a new Client.
func New(api API) *Client {
	return &Client{api: api}
}

// Get fetches a secret by ID and returns it as key-value pairs.
func (c *Client) Get(ctx context.Context, secretID string) (map[string]string, error) {
	out, err := c.api.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", secretID, err)
	}
	if out.SecretString == nil {
		return nil, fmt.Errorf("secret %q: SecretString is nil", secretID)
	}

	var kvs map[string]string
	if err := json.Unmarshal([]byte(*out.SecretString), &kvs); err != nil {
		return nil, fmt.Errorf("secret %q: unmarshal: %w", secretID, err)
	}

	return kvs, nil
}

// Put stores key-value pairs as a JSON secret.
func (c *Client) Put(ctx context.Context, secretID string, kvs map[string]string) error {
	data, err := json.Marshal(kvs)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = c.api.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretID),
		SecretString: aws.String(string(data)),
	})
	if err != nil {
		return fmt.Errorf("put secret %q: %w", secretID, err)
	}

	return nil
}
