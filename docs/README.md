# Documentation Index

This directory contains organized documentation for the TrueNAS Scale Terraform Provider.

## 📚 Documentation Structure

### API Documentation (`api/`)
Technical API information and coverage details:
- **[API_COVERAGE.md](api/API_COVERAGE.md)** - API implementation status and roadmap
- **[API_ENDPOINTS.md](api/API_ENDPOINTS.md)** - Complete API endpoint reference

### Guides (`guides/`)
How-to guides and usage documentation:
- **[VM_IP_DISCOVERY.md](guides/VM_IP_DISCOVERY.md)** - Complete guide for discovering VM IP addresses
  - MAC address export method (works for ALL VMs)
  - Guest agent query method (requires guest agent)
  - Security options and troubleshooting

### Release Notes (`releases/`)
Version history and release information:
- **[RELEASE_NOTES_v0.2.0.md](releases/RELEASE_NOTES_v0.2.0.md)** through **[RELEASE_NOTES_v0.2.19.md](releases/RELEASE_NOTES_v0.2.19.md)**
- Detailed changelog for each version
- Upgrade notes and breaking changes
- Bug fixes and enhancements

### Testing Documentation (`testing/`)

#### Test Reports (`testing/reports/`)
Test execution results and verification:
- **[TEST_REPORT_v0.2.18.md](testing/reports/TEST_REPORT_v0.2.18.md)** - v0.2.18 test results
- **[TEST_REPORT_v0.2.18_FIX_VERIFICATION.md](testing/reports/TEST_REPORT_v0.2.18_FIX_VERIFICATION.md)** - Fix verification report

#### Validation Reports (`testing/validation/`)
Comprehensive validation and quality assurance:
- **[FINAL_VALIDATION_REPORT.md](testing/validation/FINAL_VALIDATION_REPORT.md)** - Complete validation report

## 🔗 Quick Links

### Getting Started
- [Main README](../README.md) - Project overview and quick start
- [QUICKSTART.md](../QUICKSTART.md) - Getting started guide
- [KNOWN_LIMITATIONS.md](../KNOWN_LIMITATIONS.md) - Important limitations and workarounds

### Development
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [TESTING.md](../TESTING.md) - Testing guide
- [CHANGELOG.md](../CHANGELOG.md) - Complete version history

### Advanced Topics
- [KUBERNETES_MIGRATION.md](../KUBERNETES_MIGRATION.md) - Kubernetes migration guide
- [IMPORT_GUIDE.md](../IMPORT_GUIDE.md) - Resource import documentation
- [PROJECT_SUMMARY.md](../PROJECT_SUMMARY.md) - Technical implementation details

## 📖 Finding Documentation

### By Topic

**Virtual Machines:**
- VM IP discovery: [guides/VM_IP_DISCOVERY.md](guides/VM_IP_DISCOVERY.md)
- VM resource examples: [../examples/resources/truenas_vm/](../examples/resources/truenas_vm/)
- VM device management: [../examples/resources/truenas_vm_device/](../examples/resources/truenas_vm_device/)

**API Information:**
- Implementation status: [api/API_COVERAGE.md](api/API_COVERAGE.md)
- Endpoint reference: [api/API_ENDPOINTS.md](api/API_ENDPOINTS.md)

**Version History:**
- Latest release notes: [releases/RELEASE_NOTES_v0.2.19.md](releases/RELEASE_NOTES_v0.2.19.md)
- All releases: [releases/](releases/)

**Testing:**
- Test reports: [testing/reports/](testing/reports/)
- Validation reports: [testing/validation/](testing/validation/)

### By Use Case

**I want to...**
- **Get started**: See [../README.md](../README.md) and [../QUICKSTART.md](../QUICKSTART.md)
- **Discover VM IPs**: See [guides/VM_IP_DISCOVERY.md](guides/VM_IP_DISCOVERY.md)
- **Import existing resources**: See [../IMPORT_GUIDE.md](../IMPORT_GUIDE.md)
- **Understand limitations**: See [../KNOWN_LIMITATIONS.md](../KNOWN_LIMITATIONS.md)
- **Check API coverage**: See [api/API_COVERAGE.md](api/API_COVERAGE.md)
- **See what's new**: See [releases/](releases/) or [../CHANGELOG.md](../CHANGELOG.md)
- **Contribute**: See [../CONTRIBUTING.md](../CONTRIBUTING.md)
- **Run tests**: See [../TESTING.md](../TESTING.md)

## 📝 Documentation Conventions

- **File naming**: `UPPERCASE_WITH_UNDERSCORES.md` for documentation files
- **Versioned files**: Include version number in filename (e.g., `RELEASE_NOTES_v0.2.18.md`)
- **Links**: Use relative paths from the current file location
- **Examples**: Located in `../examples/` directory, organized by resource type

## 🤝 Contributing to Documentation

If you find errors or want to improve documentation:

1. Check [../CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines
2. Submit a pull request with your changes
3. Ensure all internal links work correctly after changes
4. Update this index if adding new documentation files

## 📊 Documentation Statistics

- **Guides**: 1
- **API Documentation**: 2
- **Release Notes**: 20 versions
- **Test Reports**: 2
- **Validation Reports**: 1

Last Updated: 2025-11-10