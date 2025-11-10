# Automated GitHub Releases with GPG Signing - Setup Guide

This guide provides complete instructions for setting up automated GitHub releases with GPG signing for the Terraform TrueNAS Scale provider.

## Prerequisites

Before starting, ensure you have:

1. **GPG Key Generated**: Follow [`GPG_SIGNING_SETUP.md`](./GPG_SIGNING_SETUP.md) to generate the GPG key
2. **Repository Access**: Admin access to the GitHub repository to configure secrets
3. **Public Key Added**: GPG public key added to your GitHub account (Settings → SSH and GPG keys)

## Overview

The automated release system consists of:

- **GitHub Actions Workflow**: [`.github/workflows/release.yml`](../../.github/workflows/release.yml)
- **GoReleaser Configuration**: [`.goreleaser.yml`](../../.goreleaser.yml)
- **GitHub Secrets**: Three secrets for GPG key management
- **Tag-based Triggers**: Releases are triggered by pushing version tags

## Step 1: Configure GitHub Secrets

Navigate to your repository: **Settings → Secrets and variables → Actions → New repository secret**

### Secret 1: GPG_FINGERPRINT

**Name:** `GPG_FINGERPRINT`

**Value:**
```
40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

### Secret 2: GPG_PRIVATE_KEY

**Name:** `GPG_PRIVATE_KEY`

**To get the value, run:**
```bash
gpg --armor --export-secret-keys 40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

Copy the entire output including:
```
-----BEGIN PGP PRIVATE KEY BLOCK-----
...
-----END PGP PRIVATE KEY BLOCK-----
```

⚠️ **CRITICAL**: Never commit this private key to version control or share it publicly.

### Secret 3: PASSPHRASE

**Name:** `PASSPHRASE`

**Value:** (Leave empty)

The GPG key was generated without a passphrase for CI/CD automation compatibility.

### Verification

After adding secrets, verify they appear in **Settings → Secrets and variables → Actions**:
- ✓ `GPG_FINGERPRINT`
- ✓ `GPG_PRIVATE_KEY`
- ✓ `PASSPHRASE`

## Step 2: Understand the Workflow

The GitHub Actions workflow ([`.github/workflows/release.yml`](../../.github/workflows/release.yml)) automatically:

1. **Triggers** on any tag push matching pattern `v*` (e.g., `v0.3.0`)
2. **Checks out** the repository with full git history
3. **Sets up Go** using the version specified in `go.mod`
4. **Imports GPG key** from GitHub Secrets
5. **Runs GoReleaser** to build, sign, and publish the release

### Workflow Steps Explained

```yaml
- Import GPG key: Imports the private key into the runner's GPG keyring
- Run GoReleaser: Builds binaries for multiple platforms, creates checksums, signs them, and creates the GitHub release
```

### Environment Variables

The workflow provides these environment variables to GoReleaser:
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions for creating releases
- `GPG_FINGERPRINT`: Your secret containing the GPG key fingerprint

## Step 3: Create a Release

### Using Signed Tags (Recommended)

Create and push a signed tag:

```bash
# Ensure GPG is configured locally
export GPG_FINGERPRINT=40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43

# Create signed tag
git tag -s v0.3.0 -m "Release v0.3.0"

# Push tag to trigger workflow
git push origin v0.3.0
```

### Using Regular Tags

If you prefer regular tags, the workflow will still sign the release artifacts:

```bash
git tag v0.3.0
git push origin v0.3.0
```

### Monitor the Release

1. Go to **Actions** tab in GitHub
2. Find the "Release" workflow run
3. Monitor progress through each step
4. Once complete, check the **Releases** page for the new release

## Step 4: Verify the Release

After the workflow completes:

### Check Release Assets

Navigate to **Releases** page and verify:

- ✓ Binary archives for all platforms (Linux, macOS, Windows, FreeBSD)
- ✓ `terraform-provider-truenas_X.X.X_SHA256SUMS` file
- ✓ `terraform-provider-truenas_X.X.X_SHA256SUMS.sig` signature file
- ✓ `terraform-provider-truenas_X.X.X_manifest.json` file

### Verify GPG Signature

Users can verify the signature:

```bash
# Import your public key
curl -sL https://github.com/YOUR_USERNAME.gpg | gpg --import

# Verify signature
gpg --verify terraform-provider-truenas_0.3.0_SHA256SUMS.sig \
     terraform-provider-truenas_0.3.0_SHA256SUMS
```

Expected output:
```
gpg: Good signature from "Terraform TrueNAS Provider <terraform-provider@truenas-scale.local>"
```

## Step 5: Publish to Terraform Registry

After a successful automated release:

1. **Navigate to** [Terraform Registry](https://registry.terraform.io/publish/provider/new)
2. **Select Repository**: Choose your GitHub repository
3. **Add GPG Key**: If not already added, upload your public key
4. **Publish**: The registry will automatically detect and verify the signed release

The registry validates:
- GPG signature on SHA256SUMS file
- Manifest file presence and format
- Binary archives for required platforms

## Workflow Configuration Details

### Trigger Configuration

```yaml
on:
  push:
    tags:
      - 'v*'
```

- Triggers on any tag starting with `v`
- Examples: `v0.3.0`, `v1.0.0`, `v0.3.0-beta.1`

### Permissions

```yaml
permissions:
  contents: write
```

- Allows the workflow to create GitHub releases
- Required for GoReleaser to publish release assets

### GoReleaser Action

```yaml
uses: goreleaser/goreleaser-action@v6
with:
  distribution: goreleaser
  version: latest
  args: release --clean
```

- Uses latest GoReleaser version
- `--clean` flag removes previous build artifacts before building

## Troubleshooting

### GPG Import Fails

**Error:** `gpg: no valid OpenPGP data found`

**Solution:**
- Verify `GPG_PRIVATE_KEY` includes full ASCII armor format
- Check for line break issues when copying the key
- Ensure key starts with `-----BEGIN PGP PRIVATE KEY BLOCK-----`

### Signature Verification Fails

**Error:** `gpg: Can't check signature: No public key`

**Solution:**
- Verify `GPG_FINGERPRINT` is exactly: `40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43`
- Ensure no extra spaces or characters
- Confirm public key is added to GitHub account

### Workflow Doesn't Trigger

**Solution:**
- Verify tag matches pattern `v*`
- Check tag was pushed: `git push origin v0.3.0`
- Ensure workflow file is committed to main branch
- Check workflow file syntax is valid YAML

### Release Assets Missing

**Solution:**
- Check workflow logs for errors during GoReleaser execution
- Verify `.goreleaser.yml` configuration is valid
- Ensure all required files are present (go.mod, main.go, etc.)

### Binary Build Fails

**Solution:**
- Verify `go.mod` specifies a valid Go version
- Test build locally: `go build .`
- Check for platform-specific build issues in workflow logs

## Testing the Setup

### Initial Test Release

Before creating an official release, test with a pre-release tag:

```bash
# Create test tag
git tag v0.3.0-test
git push origin v0.3.0-test

# Monitor the workflow
# If successful, delete the test release and tag
```

### Delete Test Release

```bash
# Delete remote tag
git push --delete origin v0.3.0-test

# Delete local tag
git tag -d v0.3.0-test

# Delete release from GitHub UI
```

## Security Best Practices

1. **Private Key Storage**: Only store in GitHub Secrets, never commit to repository
2. **Access Control**: Limit repository secret access to necessary team members
3. **Key Rotation**: Plan to rotate GPG keys periodically (see [`GPG_SIGNING_SETUP.md`](./GPG_SIGNING_SETUP.md))
4. **Backup**: Maintain secure backups of the private key
5. **Monitoring**: Review workflow runs for unauthorized access attempts

## Release Checklist

Before creating a release:

- [ ] Update [`CHANGELOG.md`](../../CHANGELOG.md) with version changes
- [ ] Update version references in documentation
- [ ] Run tests locally: `go test ./...`
- [ ] Test build locally: `go build .`
- [ ] Commit all changes to main branch
- [ ] Create and push signed tag
- [ ] Monitor GitHub Actions workflow
- [ ] Verify release assets are created
- [ ] Test download and installation
- [ ] Publish to Terraform Registry (if applicable)
- [ ] Announce release (GitHub Discussions, Twitter, etc.)

## Manual Release (Fallback)

If automated release fails, you can create releases manually:

```bash
# Set GPG fingerprint
export GPG_FINGERPRINT=40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43

# Run GoReleaser locally
goreleaser release --clean
```

Note: Manual releases require:
- GoReleaser installed locally
- GPG key available in local keyring
- `GITHUB_TOKEN` environment variable set

## Additional Resources

- [`GPG_SIGNING_SETUP.md`](./GPG_SIGNING_SETUP.md) - Complete GPG key setup guide
- [`GITHUB_SECRETS_SETUP.md`](./GITHUB_SECRETS_SETUP.md) - GitHub Secrets configuration
- [GoReleaser Documentation](https://goreleaser.com/) - GoReleaser reference
- [GitHub Actions Documentation](https://docs.github.com/en/actions) - GitHub Actions reference
- [Terraform Registry Publishing](https://www.terraform.io/docs/registry/providers/publishing.html) - Registry publication guide

## Support

If you encounter issues:

1. Check workflow logs in GitHub Actions tab
2. Review troubleshooting section above
3. Verify all secrets are correctly configured
4. Test GPG key locally with `gpg --list-secret-keys`
5. Open an issue with workflow logs if problems persist

## Summary

Once configured, the automated release process is:

1. **Update code** and commit to main branch
2. **Create tag**: `git tag -s vX.Y.Z -m "Release vX.Y.Z"`
3. **Push tag**: `git push origin vX.Y.Z`
4. **Wait**: GitHub Actions automatically builds, signs, and publishes
5. **Verify**: Check release page for signed artifacts
6. **Publish**: Submit to Terraform Registry if needed

The system handles:
- Multi-platform binary builds
- Checksum generation
- GPG signature creation
- GitHub release creation
- Asset uploading

All automatically, with every tag push! 🚀