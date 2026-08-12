# Release approval trust root

`clean-code release-gate` verifies a detached Ed25519 signature over an approval manifest that binds the repository, final revision, and every release evidence digest.

The trusted public key is a protected external input. It must come from organization or release authority configuration and must never be loaded from the change checkout being approved.

A release root must be a clean Git checkout at the exact full lowercase commit SHA in the signed manifest and binding.
