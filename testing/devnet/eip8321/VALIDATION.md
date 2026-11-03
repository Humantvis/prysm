# Validation record

## Live network, 2026-11-02

Built from the `augustas/devnet` working tree based on `b458aa7e5`. The local image labels include `-dirty`; the binaries used for this run have these SHA-256 digests:

| Binary | SHA-256 |
| --- | --- |
| beacon-chain | `6f6c2a52fde162924ffd8b121ca965bcaa6b85c44e4da1ca341d6acd2df79f8c` |
| validator | `16f5d431da3fca206d66017a7eaa43cd117935c85e81a48312c1a42f18901b15` |

Two Prysm nodes, two Geth execution clients and 128 validators ran with `network-params.yaml`. Evidence collected at 12:35:12 Europe/Vilnius showed:

- EIP-8321 fork epoch: 4; version: `0xe8321101`.
- Both nodes agreed on slot 60, root `0x13f3a2548d75c39264e475d018c3d9e7ed439193ae0ce26d8f58548834440053`.
- Finalized epoch: 5; `execution_optimistic: false`.
- 128 registrations and 5 nonzero hash-chain reveals, starting at slot 56.
- No missing blocks in the collected slot range 32–60.
- The independent Python verifier checked all 5 BLAKE3 links, activation delay and required BLS infinity signature.

Raw evidence remains outside Git in the workspace's `.tools/devnet-evidence/randao-live-20261102/`. It contains public block JSON, `summary.json` and `hash-verification.json`. Private validator keys, chain seeds, databases and Docker disks are not part of this change. Re-run the collection and verification commands in README.md to produce fresh evidence.

The earlier run also checked 70 links through slot 125, but used the images before the historical state-diff fix. The figures above refer to the latest images including that fix.

## Bazel checks

The affected mainnet-preset suites passed:

- `//api/server/structs:go_default_test`
- `//beacon-chain/execution:go_default_test`
- `//beacon-chain/p2p:go_default_test`
- `//beacon-chain/p2p/types:go_default_test`
- `//validator/client:go_default_test`
- `//beacon-chain/db/kv:go_default_test`
- `//consensus-types/hdiff:go_default_test`

The database and hdiff suites were rerun after fixing EIP-8321 state-diff persistence and both passed. Focused minimal-preset validator RPC tests cover EIP-8321 generic block construction and commitment activation across an epoch boundary without mutating the canonical head state.

These results do not mean every Prysm test passed. The full minimal-preset validator/client suite has an upstream bitvector-size compilation mismatch; that suite was tested using the mainnet preset. The live network uses the minimal preset. Cross-client interoperability, complete REST publication/debug-state APIs and production readiness remain outside this validation.
