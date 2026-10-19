# EIP-8321 foundation on Heze

This branch targets `eip-8321-heze`. EIP-8321 is a separate feature fork after
Heze, following `upgrade_to_eip8321`; the existing Gloas and Heze layouts stay
unchanged.

## Specification

The reference is `ethereum/consensus-specs` commit
`460fd37c3a58f6baf8a9a285d2adf71c4f818929`:

- `specs/_features/eip8321/beacon-chain.md`
- `specs/_features/eip8321/fork.md`
- `configs/mainnet.yaml` and `configs/minimal.yaml`

Registration domain: `0x11000000`. Registration delay: 3 epochs. Maximum
registrations processed per block: 128. The registry, pending queue, and block
registration list all use ProgressiveList; 128 is a processing limit, not an SSZ
list bound. The decoder retains the generator's finite allocation guards (2^40
for the three new lists), independently of their progressive Merkle roots.
The EIP-8321 state has 44 active fields at 48 stable positions; its body has 13
active fields at 15 positions. Removed Heze positions remain zero leaves.

`ContainerRootProgressive` now preserves those gaps instead of compacting active
roots. Its sparse-container regression uses an independent golden root from
`eth-ssz-specs==0.1.0`, the Python library used by the pinned spec generator.

Fork versions are `0xe8321000` (mainnet preset) and `0xe8321001` (minimal).
Other built-in configurations use distinct local feature versions to avoid
config registry collisions. Activation defaults to the far-future epoch.

## Team interfaces

New protobuf and generated SSZ types are defined in `eip8321.proto` and
`eip8321.yaml`. Native state initialization uses
`InitializeFromProtoEip8321` / `InitializeFromProtoUnsafeEip8321`.

- `RandaoCommitmentAtIndex`, `RandaoCommitments`, `PendingRandaoCommitments`
- `SetRandaoCommitments`, `UpdateRandaoCommitmentAtIndex`, `AppendRandaoCommitment`
- `SetPendingRandaoCommitments`, `AppendPendingRandaoCommitment`
- Block body: `HashChainReveal`, `RandaoCommitmentRegistrations`
- Signed block: `SetHashChainReveal`, `SetRandaoCommitmentRegistrations`
- `crypto/hash.Blake3` and `config/fieldparams.HashChainRandaoDST`
- `core/eip8321.UpgradeToEip8321` and `runtime/version.Eip8321`

The upgrade preserves all Heze fields, updates fork metadata, creates a zero
commitment for every validator, and clears the pending queue. Validator
registration through `AddValidatorToRegistry` appends a zero commitment on the
feature fork.

Consensus processing (registrations, activation, hash-chain RANDAO, signatures),
networking, and validator production flows belong to the remaining team work.
Those paths must use the EIP-8321 version and protobuf types rather than adding
fields to Heze itself. This foundation does not claim full EIP interoperability.

### Validator RPC integration

The validator gRPC service also defines:

- `SubmitRandaoCommitmentRegistration(SignedRandaoCommitmentRegistration) -> google.protobuf.Empty`
- `RandaoCommitment(ValidatorIndexRequest) -> RandaoCommitmentResponse`

The query uses `ValidatorIndexRequest.public_key`; the response contains a
32-byte `commitment`. A zero commitment denotes an unregistered validator.
`SignRequest.randao_commitment` uses field 131; `block_eip8321` retains field 130.
`BlockRequest.hash_chain_reveal` remains field 8.

The server embeds the generated unimplemented service so this schema change
builds before the validator implementation is merged. Its new RPCs return
`Unimplemented` until the concrete methods from the validator integration are
present. Merge the `.proto` sources first and regenerate code with `make gen`
instead of resolving generated protobuf and mock conflicts by hand.

Validation on 2026-10-19: mainnet protobuf and EIP-8321 package tests passed,
minimal protobuf tests passed, and all four binaries built with Bazel. The new
tests exercise both RPCs over in-memory gRPC and serialize the signing payload.

An isolated integration of `eip8321-heze/matas` at
`73f8f7f2c2c14b792fc8cecf2288dc2981b9f05a` with shared base `efbbc5391`
and these protobuf changes still fails outside the protobuf layer:

- `validator/client/randao_chain.go:62`: `rand.Bytes` does not exist in the
  imported project package. Use a supported secure random source.
- `beacon-chain/sync/subscriber_randao_commitment_registration.go`: the `log`
  import conflicts with the package logger, and `operation` is unused.
- `beacon-chain/operations/randaocommitments/pool.go:91`: the `uintcast`
  analyzer rejects an unchecked `uint64` to `int` conversion.

Separately, `validator/randao_commitment.go` under the beacon RPC package calls
`Broadcast(ctx, topic, req)`, while `p2p.Broadcaster.Broadcast` accepts only
`(context.Context, proto.Message)`. The topic is selected by message type.
These are validator integration follow-ups; this validation does not claim that
the complete Matas branch builds. Regenerate its interface mocks after merging
the shared base and protobuf changes, then rerun the build and tests.

## Generate and run reference vectors

Use Linux or WSL with the consensus-specs prerequisites (Python 3.13, uv, make):

```sh
git clone https://github.com/ethereum/consensus-specs.git
cd consensus-specs
git checkout 460fd37c3a58f6baf8a9a285d2adf71c4f818929
make test reftests=true fork=eip8321
```

The output is `reftests/`. To generate only the foundation's fork cases sooner:

```sh
make test reftests=true fork=eip8321 k=eip8321_fork_basic
```

That subset generates 7 minimal and 6 mainnet upgrade cases. The Go runner
compares upgraded SSZ bytes against each generated post-state and compares the
native and generated state roots. It fails if the supplied directory contains
no matching fork cases. Without the environment variable, only this optional
reference test skips; the regular unit tests still run.

From the Prysm checkout, expose the generated directory to Bazel:

```sh
bazel test //beacon-chain/core/eip8321:go_default_test \
  --test_env=EIP8321_SPEC_TESTS_DIR=/absolute/path/to/consensus-specs/reftests \
  --test_strategy=standalone --test_output=errors
bazel test //beacon-chain/core/eip8321:go_default_test --config=minimal \
  --test_env=EIP8321_SPEC_TESTS_DIR=/absolute/path/to/consensus-specs/reftests \
  --test_strategy=standalone --test_output=errors
```

For Docker, mount the vector directory read-only and use its container path.
Keep the generated dataset outside the repository; commit the spec revision and
generation instructions instead.

## Validation on 2026-10-17

- The focused Python fork generator completed: 13 passed, 8 skipped.
- Go passed all 13 generated upgrades: 6 mainnet and 7 minimal cases.
- Eight mainnet Bazel test targets passed: EIP-8321, native state, blocks,
  configuration, BLAKE3, SSZ, fork detection, and runtime versions.
- Three minimal Bazel test targets passed: EIP-8321, blocks, and protobuf/SSZ.
- Full code generation, Go formatting/imports, and Gazelle checks passed.
- Bazel built `beacon-chain`, `validator`, `prysmctl`, and `client-stats`.

The broad Python generation was stopped after 435 passing cases. This is not a
claim that the full EIP-8321 consensus suite passed. Windows checkout tests used
a local Linux fixture-directory adapter; the reference runner itself is standard
Bazel/Go and accepts the generated vector directory through the environment.
