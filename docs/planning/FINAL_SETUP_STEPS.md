# Final Setup Steps for Automated GitHub Releases

This document provides the essential steps to complete the automated release setup.

## ✅ What's Already Done

- [x] GitHub Actions workflow created ([`.github/workflows/release.yml`](../../.github/workflows/release.yml))
- [x] GoReleaser configuration verified ([`.goreleaser.yml`](../../.goreleaser.yml))
- [x] GPG key generated (fingerprint: `40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43`)
- [x] Documentation created

## 🔧 Required Setup Steps

### Step 1: Export GPG Private Key

Run this command to export your GPG private key:

```bash
gpg --armor --export-secret-keys 40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

**Save the output** - you'll need it for GitHub Secrets.

### Step 2: Configure GitHub Secrets

Go to: **Repository Settings → Secrets and variables → Actions → New repository secret**

Add these three secrets:

#### 1. GPG_FINGERPRINT
```
40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

#### 2. GPG_PRIVATE_KEY
Paste the complete output from Step 1, including:
```
-----BEGIN PGP PRIVATE KEY BLOCK-----
...
-----END PGP PRIVATE KEY BLOCK-----
```

#### 3. PASSPHRASE
Leave empty (or set to empty string)

### Step 3: Verify Setup

Check that all secrets are configured:
- ✓ `GPG_FINGERPRINT`
- ✓ `GPG_PRIVATE_KEY`  
- ✓ `PASSPHRASE`

### Step 4: Test the Workflow

Create a test release:

```bash
# Create test tag
git tag v0.3.0-test
git push origin v0.3.0-test
```

Monitor the workflow in the **Actions** tab. If successful, you'll see:
- ✓ Checkout completed
- ✓ Go setup completed
- ✓ GPG key imported
- ✓ GoReleaser run completed
- ✓ Release created with signed artifacts

### Step 5: Clean Up Test Release (Optional)

If the test was successful, clean up:

```bash
# Delete remote tag
git push --delete origin v0.3.0-test

# Delete local tag
git tag -d v0.3.0-test
```

Then manually delete the release from GitHub UI.

## 🚀 Creating Real Releases

Once setup is complete, create releases with:

```bash
# Create signed tag
git tag -s v0.3.0 -m "Release v0.3.0"

# Push tag to trigger automated release
git push origin v0.3.0
```

The workflow will automatically:
1. Build binaries for all platforms
2. Generate checksums
3. Sign with GPG
4. Create GitHub release
5. Upload all artifacts

## 📚 Full Documentation

For detailed information, see:

- [`AUTOMATED_RELEASE_SETUP.md`](./AUTOMATED_RELEASE_SETUP.md) - Complete setup guide
- [`GPG_SIGNING_SETUP.md`](./GPG_SIGNING_SETUP.md) - GPG key management
- [`GITHUB_SECRETS_SETUP.md`](./GITHUB_SECRETS_SETUP.md) - Secrets configuration

## 🆘 Troubleshooting

If the workflow fails:

1. **Check workflow logs** in Actions tab
2. **Verify secrets** are correctly configured
3. **Confirm GPG key** is available locally with: `gpg --list-secret-keys`
4. **Review** [`AUTOMATED_RELEASE_SETUP.md`](./AUTOMATED_RELEASE_SETUP.md) troubleshooting section

## 🔒 Security Reminders

- ❌ Never commit the private GPG key to version control
- ✓ Keep private key backup in secure location
- ✓ Limit access to repository secrets
- ✓ Rotate keys periodically

## ✨ Next Steps After Setup

1. Update [`CHANGELOG.md`](../../CHANGELOG.md) for next release
2. Test release workflow with a test tag
3. Create official release when ready
4. Publish to Terraform Registry if needed

---

**Need Help?** See the full documentation or open an issue.