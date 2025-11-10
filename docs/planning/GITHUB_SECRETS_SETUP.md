# GitHub Secrets Configuration for GPG Signing

This document provides the exact values needed for GitHub Secrets to enable automated GPG signing in releases.

## Required GitHub Secrets

Configure these in your repository: **Settings → Secrets and variables → Actions → New repository secret**

### 1. GPG_FINGERPRINT

**Name:** `GPG_FINGERPRINT`

**Value:**
```
40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

This is the full 40-character fingerprint of the GPG key.

---

### 2. GPG_PRIVATE_KEY

**Name:** `GPG_PRIVATE_KEY`

**Value:** You need to export the private key in ASCII armor format.

**To get this value, run:**

```bash
gpg --armor --export-secret-keys 40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43
```

This will output the private key starting with:
```
-----BEGIN PGP PRIVATE KEY BLOCK-----
...
-----END PGP PRIVATE KEY BLOCK-----
```

**Copy the entire output** (including the BEGIN and END lines) and paste it as the secret value.

⚠️ **Security Note:** This private key should NEVER be committed to version control or shared publicly.

---

### 3. PASSPHRASE

**Name:** `PASSPHRASE`

**Value:** (Leave empty or set to empty string)

This key was generated without a passphrase for CI/CD automation compatibility, so this secret should be empty or not set at all.

---

## Verification

After adding these secrets, they should appear in:
**Settings → Secrets and variables → Actions → Repository secrets**

You should see:
- ✓ `GPG_FINGERPRINT`
- ✓ `GPG_PRIVATE_KEY`
- ✓ `PASSPHRASE` (optional)

## Next Steps

Once these secrets are configured:

1. **Create a GitHub Actions workflow** (if not already done) - see [`.github/workflows/release.yml`](../../.github/workflows/release.yml) example in the main documentation

2. **Test the setup** by creating a test release:
   ```bash
   git tag -s v0.3.0-test -m "Test release"
   git push origin v0.3.0-test
   ```

3. **Monitor the Actions tab** in GitHub to verify the release workflow runs successfully and signs the artifacts

## Troubleshooting

### GPG import fails in Actions

If you see errors like "gpg: no valid OpenPGP data found":
- Ensure the `GPG_PRIVATE_KEY` includes the full ASCII armor format
- Check for any line break issues when copying the key
- Verify the key starts with `-----BEGIN PGP PRIVATE KEY BLOCK-----`

### Signature verification fails

- Confirm `GPG_FINGERPRINT` matches exactly: `40DF8713FDA8BC549BF2BB6E83353AF4D56A8F43`
- Ensure no extra spaces or characters in the fingerprint
- Verify the public key is added to your GitHub account

## Additional Resources

See [`GPG_SIGNING_SETUP.md`](./GPG_SIGNING_SETUP.md) for complete documentation on the GPG signing infrastructure.