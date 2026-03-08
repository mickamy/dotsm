package cmd_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type stubAPI struct {
	secret    string
	putCalled bool
}

func (s *stubAPI) GetSecretValue(
	_ context.Context,
	_ *secretsmanager.GetSecretValueInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(s.secret),
	}, nil
}

func (s *stubAPI) PutSecretValue(
	_ context.Context,
	_ *secretsmanager.PutSecretValueInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.PutSecretValueOutput, error) {
	s.putCalled = true
	return &secretsmanager.PutSecretValueOutput{}, nil
}
