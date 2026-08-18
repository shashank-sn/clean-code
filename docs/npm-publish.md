# Publishing to npm

**Live package:** https://www.npmjs.com/package/@shashanksn/clean-code

The unscoped name `clean-code` is blocked by npm (similar to `cleancode`). The scoped package installs the `clean-code` CLI binary.

```bash
npm install -g @shashanksn/clean-code
clean-code version
```

Deprecated predecessor: `clean-code-skills` → use `@shashanksn/clean-code`.

## Publish

From repository root:

```bash
npm login
npm publish --access public
```

`postinstall` downloads Go 1.22 when it is not already on PATH (Node is required to run npm). The one-line `install.sh` bootstrap installs both Node 20 and Go when missing.
