### Fixed

- EIP-8321: validators now propose with a BLS randao reveal before the EIP-8321 fork, instead of missing every proposal when the fork is not scheduled.
- EIP-8321: network configs without `EIP8321_FORK_VERSION` no longer fail to load with a fork version conflict against the built-in minimal/mainnet config.
