package sm_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/mickamy/dotsm/internal/sm"
)

type getFunc = func(
	ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error)

type putFunc = func(
	ctx context.Context,
	params *secretsmanager.PutSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.PutSecretValueOutput, error)

type mockAPI struct {
	getFunc getFunc
	putFunc putFunc
}

func (m *mockAPI) GetSecretValue(
	ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	return m.getFunc(ctx, params, optFns...)
}

func (m *mockAPI) PutSecretValue(
	ctx context.Context,
	params *secretsmanager.PutSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.PutSecretValueOutput, error) {
	return m.putFunc(ctx, params, optFns...)
}

func TestClient_Get(t *testing.T) {
	t.Parallel()

	//nolint:gosec // test data, not real credentials
	tests := []struct {
		name    string
		secret  string
		want    map[string]string
		wantErr bool
	}{
		{
			name:   "valid JSON secret",
			secret: `{"HOST":"localhost","PORT":"5432"}`,
			want:   map[string]string{"HOST": "localhost", "PORT": "5432"},
		},
		{
			name:    "invalid JSON",
			secret:  `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			api := &mockAPI{
				getFunc: func(
					_ context.Context,
					params *secretsmanager.GetSecretValueInput,
					_ ...func(*secretsmanager.Options),
				) (*secretsmanager.GetSecretValueOutput, error) {
					if got := *params.SecretId; got != "test/app" {
						t.Errorf("secret ID: got %q, want %q", got, "test/app")
					}
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(tt.secret),
					}, nil
				},
			}

			client := sm.New(api)
			got, err := client.Get(ctx, "test/app")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("key count: got %d, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("key %q: got %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

func TestClient_Get_NilSecretString(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	api := &mockAPI{
		getFunc: func(
			_ context.Context,
			_ *secretsmanager.GetSecretValueInput,
			_ ...func(*secretsmanager.Options),
		) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{SecretString: nil}, nil
		},
	}

	client := sm.New(api)
	_, err := client.Get(ctx, "test/app")
	if err == nil {
		t.Fatal("expected error for nil SecretString")
	}
}

func TestClient_Put(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	kvs := map[string]string{"FOO": "bar", "BAZ": "qux"}

	var captured string
	api := &mockAPI{
		putFunc: func(
			_ context.Context,
			params *secretsmanager.PutSecretValueInput,
			_ ...func(*secretsmanager.Options),
		) (*secretsmanager.PutSecretValueOutput, error) {
			if got := *params.SecretId; got != "test/app" {
				t.Errorf("secret ID: got %q, want %q", got, "test/app")
			}
			captured = *params.SecretString
			return &secretsmanager.PutSecretValueOutput{}, nil
		},
	}

	client := sm.New(api)
	if err := client.Put(ctx, "test/app", kvs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(captured), &got); err != nil {
		t.Fatalf("unmarshal captured: %v", err)
	}
	for k, wantV := range kvs {
		if got[k] != wantV {
			t.Errorf("key %q: got %q, want %q", k, got[k], wantV)
		}
	}
}
