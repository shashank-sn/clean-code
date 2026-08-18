# Publishing to npm

**Live package:** https://www.npmjs.com/package/@shashanksn/clean-code

The unscoped name `clean-code` is blocked by npm (similar to `cleancode`). The scoped package installs the `clean-code` CLI binary.

```bash
npm install -g @shashanksn/clean-code
clean-code version
```

Deprecated predecessor: `clean-code-skills` → use `@shashanksn/clean-code`.

## Automated publish (recommended)

1. Add repository secret **`NPM_TOKEN`** (npm granular token with publish access for `@shashanksn/clean-code`).
2. Bump `package.json` `version` and update `CHANGELOG.md`.
3. Push a tag matching the version:

```bash
git tag v0.2.4
git push origin v0.2.4
```

4. GitHub Actions:
   - `release.yml` — builds platform binaries and creates a GitHub Release with artifacts.
   - `npm-publish.yml` — runs on `release: published`, tests, and `npm publish --provenance --access public`.

Manual re-run: Actions → **npm-publish** → **Run workflow**.

## Manual publish

From repository root:

```bash
npm login
npm publish --access public
```

Scoped packages require `--access public` for global install without npm login to the scope.

Go bootstrap and native CLI build happen on first `clean-code` run (no `postinstall` script).
