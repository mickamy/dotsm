# dotsm

Sync AWS Secrets Manager secrets with `.env` files — pull, push, and diff in a single binary.

## Install

### Homebrew

```bash
brew install mickamy/tap/dotsm
```

### Go

```bash
go install github.com/mickamy/dotsm@latest
```

### Binary

Download from [Releases](https://github.com/mickamy/dotsm/releases).

## Usage

### Pull

Fetch a secret and write as `.env`:

```bash
# Write to file
dotsm pull -secret myapp-prod/app -output .env

# Print to stdout
dotsm pull -secret myapp-prod/app
```

### Push

Read a `.env` file and store as a JSON secret:

```bash
dotsm push -secret myapp-prod/app -input .env

# Preview without writing
dotsm push -secret myapp-prod/app -input .env -dry-run
```

### Diff

Compare local `.env` with remote secret:

```bash
dotsm diff -secret myapp-prod/app -input .env
```

Output:

```
+ NEW_KEY          # in local, not in remote
- REMOVED_KEY      # in remote, not in local
~ CHANGED_KEY (changed)
```

Exits with code 1 if differences are found — useful in CI.
Secret values are masked by default. Use `-show-values` to reveal them.

## Common Options

| Flag       | Description                             |
|------------|-----------------------------------------|
| `-secret`  | Secret ID in Secrets Manager (required) |
| `-region`  | AWS region (overrides default)          |
| `-profile` | AWS CLI profile                         |

## Prerequisites

The `push` command updates an **existing** secret. Create the secret beforehand (e.g., via Terraform/OpenTofu or AWS CLI):

```bash
aws secretsmanager create-secret --name myapp-prod/app --secret-string '{}'
```

## Limitations

- Multiline values are escaped as `\n` / `\r` in `.env` output. Raw multiline `.env` syntax is not supported.
- Secrets must be JSON objects with string values (`{"KEY": "value"}`).

## AWS Authentication

dotsm uses the standard AWS SDK credential chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (EC2, ECS, Lambda)

## License

[MIT](./LICENSE)
