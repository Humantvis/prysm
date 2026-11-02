### Added

- Reproducible two-node Prysm EIP-8321 development network with a pinned ethereum-package patch, Docker build helpers and live evidence collection.
- Automatic validator RANDAO registration with persistent hash-chain seeds and epoch retries.
- EIP-8321 block JSON responses, gossip fork mappings and database persistence.

### Fixed

- Route EIP-specific validator calls through gRPC while retaining REST for Heze duties.
- Account for commitment activation at epoch boundaries and retain the Heze execution-engine API after the feature fork.
- Preserve active and pending RANDAO commitments when serializing and applying historical state diffs, including the Heze-to-EIP-8321 upgrade.
