# Running the EIP-8321 spec tests

The EIP-8321 vectors are not part of the released consensus-spec-tests archive,
so they are generated locally and linked into Prysm's test-data directory.

## 1. Generate the vectors (consensus-specs)

```sh
git clone https://github.com/ethereum/consensus-specs.git
cd consensus-specs
make test reftests=true fork=eip8321
```

The output is a `reftests/` directory with `minimal/eip8321/...` and
`mainnet/eip8321/...`.

## 2. Link them into Prysm

From the Prysm checkout (`third_party/testdata` is git-ignored):

```sh
mkdir -p third_party/testdata/tests/minimal third_party/testdata/tests/mainnet
ln -sfn /absolute/path/to/reftests/minimal/eip8321 third_party/testdata/tests/minimal/eip8321
ln -sfn /absolute/path/to/reftests/mainnet/eip8321 third_party/testdata/tests/mainnet/eip8321
```

## 3. Run

```sh
go test -tags minimal ./testing/spectest/minimal/eip8321/ -v
go test ./testing/spectest/mainnet/eip8321/ -v
```

| Handler | Runner |
|---|---|
| operations/randao | `blocks.ProcessRandao` (EIP-8321 branch) |
| operations/randao_commitment_registration | `eip8321.ProcessRandaoCommitmentRegistrations` |
| epoch_processing/pending_randao_commitments | `eip8321.ProcessPendingRandaoCommitments` |
| epoch_processing/randao_mixes_reset | `electra.ProcessRandaoMixesReset` |
| fork/fork | `eip8321.UpgradeToEip8321` |
| sanity/blocks (minimal only) | full state transition |

`operations/randao/hash_chain_reveal_consumes_one_link_per_block` is skipped:
the generator writes only a manifest for that multi-block case. The mainnet
`sanity/blocks` lifecycle case is not generated upstream (it needs 256
registrations in one block, above the limit of 128).
