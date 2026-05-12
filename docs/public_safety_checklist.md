# Public Safety Checklist

Run this checklist before the first public push and before every release tag.

## Repository Hygiene

- Current local module path is a release-preparation placeholder, not a public tag path.
- Before publication, choose a controlled remote/module path, add `LICENSE`, remove any `replace` directives, and rerun clean checks.
- No local `replace` directives pointing to private or sibling checkouts.
- No generated binaries, deployment bundles, package archives, or runtime state.
- No private screenshots, captures, DBC files, customer procedures, or lab
  manuals.
- No private hostnames, IPs, serials, usernames, credentials, tunnel names,
  machine names, or route defaults.

## Code And Fixture Neutrality

- Package names describe public concepts, not migration history.
- Tests use fictional fixtures or public standards-inspired examples.
- Hardware-specific adapters are interfaces unless the implementation is fully
  public and vendor-general.
- Live authority examples are deterministic mocks unless a downstream private
  application supplies real adapters.

## Downstream Compatibility

- Public consumers build from fresh clones using public SignalForge
  dependencies.
- Private integrations are verified outside this public release gate.
