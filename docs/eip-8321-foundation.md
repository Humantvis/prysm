# EIP-8321 foundation handoff

This branch provides the shared types and state access needed for the team's
EIP-8321 implementation. It is based on `eip-8321` at
`cdaa03be567bea04fd7efb83b0dc6745c75019e0` and extends Gloas.

The repository also has `eip-8321-heze` and `mantvydas/randao-heze` branches.
The changes here have not been ported to Heze or tested with those branches.
They must be adapted before integration into the Heze implementation.

## Available functionality

- Configuration: registration domain `0x0f000000`, registration delay of three
  epochs, maximum of 128 registrations per block, and `HASH_CHAIN_RANDAO_DST`.
- `crypto/hash.Blake3`, checked against five official BLAKE3 test vectors.
- Registration, signed registration, and pending registration Proto/SSZ types.
- Gloas block reveal and registration fields, block interfaces, copying, Proto
  conversion, SSZ serialization, and Merkle proof roots.
- Gloas state commitment registry and progressive pending queue, state interfaces,
  independent copies, serialization, and state root updates.
- Zero commitments on Gloas upgrade and when adding a validator to the registry.
- `BlockRequest.hash_chain_reveal`, using protobuf field number 8.

State read methods are `RandaoCommitmentAtIndex`, `RandaoCommitments`, and
`PendingRandaoCommitments`. Write methods are `SetRandaoCommitments`,
`UpdateRandaoCommitmentAtIndex`, `AppendRandaoCommitment`,
`SetPendingRandaoCommitments`, and `AppendPendingRandaoCommitment`.

The pending queue uses progressive SSZ merkleization. Its generator-required
`ssz-max` annotation uses `VALIDATOR_REGISTRY_LIMIT`: reachable valid states have
at most one pending registration per validator.

## Remaining team integration

- Implement registration signatures, one-time registration checks, pending
  activation, and hash-chain reveal verification and mixing.
- Populate the block request field from the validator and pass it through block
  production and the relevant APIs; normalize an absent legacy request field to
  the protocol's zero reveal where appropriate.
- Connect operation selection, gossip, validator chain storage, and proposal
  behavior as required by the team's task split.
- Port the foundation to Heze if that is the team's chosen integration base.

## Verified locally

Validation used Linux Docker, Go 1.26.5, Bazel 7.4.1, and real cryptography.

- All four program targets built successfully: `//cmd/beacon-chain:beacon-chain`,
  `//cmd/validator:validator`, `//cmd/prysmctl:prysmctl`, and
  `//cmd/client-stats:client-stats`. The build was repeated after the final Proto
  state copying fix.
- Full Proto tests passed for mainnet and `--config=minimal`.
- Block tests passed for mainnet and minimal; mainnet block tests were repeated
  during final validation.
- Hash, hash/HTR, config/params, and config/fieldparams tests passed.
- Full state-native, state-native/custom-types, Gloas core, and Electra core
  tests passed during state integration.
- Selected RANDAO state tests also passed with the minimal preset.
- Formatting/import checks, Gazelle dependency/build checks, and all official
  generator stages passed. The generator ran in a separate Linux copy; the
  compile smoke check used Bazel rather than the Makefile's Go build path.

Use the repository's Bazel test flags:

```sh
bazel test //proto/prysm/v1alpha1:go_default_test \
  --keep_going --test_output=errors --flaky_test_attempts=3 --build_tests_only
bazel test //proto/prysm/v1alpha1:go_default_test --config=minimal \
  --keep_going --test_output=errors --flaky_test_attempts=3 --build_tests_only
```

The full repository test suite, cross-client interoperability, and end-to-end
hash-chain proposals have not been verified. Shared foundation tests do not
establish completeness of the full EIP implementation.
