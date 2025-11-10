# GPG Signing Setup for Terraform Registry

This document explains the GPG key setup for signing Terraform provider releases for the HashiCorp Registry.

## Overview

The Terraform Registry requires all provider releases to be signed with a GPG key. This ensures the authenticity and integrity of published providers.

## Key Information

**Generated Key Details:**
- Key Type: RSA 4096-bit
- Key ID: `83353AF4D56A8F43`
- Full Fingerprint: `40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43`
- Identity: `Terraform TrueNAS Provider <terraform-provider@truenas-scale.local>`
- Expiration: No expiration

## GitHub Configuration

### 1. GitHub GPG Key

The public key has been added to the GitHub account under Settings → SSH and GPG keys. This allows GitHub to verify signed commits and tags.

### 2. GitHub Secrets

For automated releases via GitHub Actions, configure these secrets in your repository (Settings → Secrets and variables → Actions):

#### Required Secrets:

- **`GPG_PRIVATE_KEY`**: The private GPG key in ASCII armor format
  ```bash
  # Export private key
  gpg --armor --export-secret-keys 40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
  ```

- **`GPG_FINGERPRINT`**: The key fingerprint
  ```
  40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
  ```

- **`PASSPHRASE`**: Empty (key was generated without passphrase for CI/CD automation)

## GoReleaser Configuration

The [`.goreleaser.yml`](../../.goreleaser.yml) is already configured for GPG signing:

```yaml
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"
```

### Environment Variables Required

When running GoReleaser locally or in CI/CD:

```bash
export GPG_FINGERPRINT=40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

## GitHub Actions Workflow

Example workflow configuration for automated releases:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
          cache: true

      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ secrets.GPG_FINGERPRINT }}
```

## Manual Release Process

For manual releases:

1. **Ensure GPG key is available:**
   ```bash
   gpg --list-secret-keys --keyid-format LONG
   ```

2. **Set environment variable:**
   ```bash
   export GPG_FINGERPRINT=40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
   ```

3. **Create signed tag:**
   ```bash
   git tag -s v0.3.0 -m "Release v0.3.0"
   git push origin v0.3.0
   ```

4. **Run GoReleaser:**
   ```bash
   goreleaser release --clean
   ```

## Verifying Signatures

Users can verify release signatures:

```bash
# Import the public key from GitHub
curl -sL https://github.com/<username>.gpg | gpg --import

# Verify the signature
gpg --verify terraform-provider-truenas_0.3.0_SHA256SUMS.sig \
     terraform-provider-truenas_0.3.0_SHA256SUMS
```

## Terraform Registry Configuration

When publishing to the Terraform Registry:

1. **Navigate to**: https://registry.terraform.io/publish/provider/new
2. **Add GPG Public Key**: Upload the ASCII-armored public key
3. The registry will use this key to verify all release signatures

## Key Management

### Backup Recommendations

**Important**: Keep secure backups of the private key:

```bash
# Export private key to secure location
gpg --armor --export-secret-keys 40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43 > gpg-private-key.asc

# Store in secure password manager or encrypted vault
```

### Key Rotation

If the key needs to be rotated:

1. Generate new GPG key with same process
2. Update GitHub secrets
3. Add new public key to GitHub account
4. Update Terraform Registry with new public key
5. Revoke old key after transition period

## Troubleshooting

### "No secret key" error

```bash
# List available secret keys
gpg --list-secret-keys --keyid-format LONG

# If key is missing, import from backup
gpg --import gpg-private-key.asc
```

### GPG not found in CI/CD

Ensure the GitHub Action workflow includes the `crazy-max/ghaction-import-gpg` step to import the key before running GoReleaser.

### Signature verification fails

Check that:
- The correct GPG_FINGERPRINT environment variable is set
- The GPG key is imported and available
- GoReleaser has access to the key

## Security Notes

- **Private Key**: Never commit the private key to version control
- **Key File**: The `terraform-registry-public-key.asc` is in `.gitignore`
- **Passphrase**: Key was generated without passphrase for automation compatibility
- **GitHub Secrets**: Private key stored as encrypted secret in GitHub
- **Access Control**: Limit access to repository secrets to necessary team members

## References

- [HashiCorp Provider Publishing](https://www.terraform.io/docs/registry/providers/publishing.html)
- [GoReleaser GPG Signing](https://goreleaser.com/customization/sign/)
- [GitHub GPG Keys](https://docs.github.com/en/authentication/managing-commit-signature-verification)