# Publishing `clean-code-skills` to npm

The package is configured but **not yet published**. Users installing from npm today will see `404 Not Found`.

## First-time publish (maintainer)

1. Create an npm account and log in: `npm login`
2. From a clean clone on the release branch:

```bash
npm publish --access public --provenance
```

3. Verify:

```bash
npm view clean-code-skills version
npm install -g clean-code-skills
clean-code version
```

## Automated publish on GitHub Release

1. Add repository secret `NPM_TOKEN` (npm access token with publish permission).
2. Create a GitHub Release (tag e.g. `v0.2.0`).
3. Workflow `.github/workflows/npm-publish.yml` runs `npm publish`.

Ensure `package.json` `version` matches the release before tagging.

## Until npm is live

Users should install from Git:

```bash
npm install -g "git+https://github.com/shashank-stitch/clean-code.git#cursor/ce-parity-shipping-pipeline-70e7"
```

Or use `scripts/install.sh` (Go only, no Node required).
