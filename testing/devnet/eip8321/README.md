# EIP-8321 Prysm demonstration

This development network runs two Prysm beacon nodes, two validator clients with 64 keys each, two Geth execution clients and Dora. It starts at Heze with FOCIL enabled and schedules the separate EIP-8321 feature fork at epoch 4. The minimal preset has 8 slots per epoch and 6 seconds per slot. These values are specific to this demonstration.

## Requirements

- Docker Desktop running Linux containers, with sufficient disk space (allow 50 GB for a fresh build).
- PowerShell 7, Git and Kurtosis CLI 1.22.2.
- Internet access for the initial dependencies and images.

The build uses Bazel 7.4.1 inside a pinned Go container. It never mounts the Docker socket into a build container. The local Kurtosis CLI controls Docker directly.

## Build and start

From the Prysm repository root:

```powershell
./testing/devnet/eip8321/Build.ps1
./testing/devnet/eip8321/Start.ps1 -KurtosisPath C:/path/to/kurtosis.exe
```

On Augustas's machine the default Kurtosis path is the sibling workspace's `.tools/kurtosis/kurtosis.exe`; `Build.ps1 -DockerPath 'C:/Program Files/Docker/Docker/resources/bin/docker.exe'` works when Docker is absent from PATH.

`Prepare-Package.ps1` creates a non-hidden sibling `ethereum-package` directory, pins revision `5795b77e65235ad223d4d3b413b2739190bf60cc`, normalizes text line endings and applies the included patch. The patch adds `eip8321_fork_epoch` / `eip8321_fork_version`, appends them to the generated consensus config, and enables the RANDAO gRPC connection alongside the REST interface required by Heze duties. Existing conflicting package changes are rejected.

The genesis generator still produces an official Heze genesis state. Prysm performs the actual feature-fork upgrade later; no EIP genesis state or replacement test results are injected.

Images:

| Component | Image |
| --- | --- |
| Our beacon node | `bct/prysm-beacon:eip8321-minimal` |
| Our validator | `bct/prysm-validator:eip8321-minimal` |
| Compatible execution client | `ethpandaops/geth:heze-devnet-2-66af2fe` |
| Upstream beacon baseline | `ethpandaops/prysm-beacon-chain:heze-devnet-2-minimal-3cc4d4e` |
| Upstream validator baseline | `ethpandaops/prysm-validator:heze-devnet-2-minimal-3cc4d4e` |

The baseline config is `heze-upstream.yaml`. It uses the upstream images and no EIP fork. The modified network is `network-params.yaml`, with unique feature version `0xe8321101`. Image labels identify the source revision, with `-dirty` when built from uncommitted changes. Binary hashes are saved alongside the exported binaries.

## Demonstration

```powershell
./testing/devnet/eip8321/Status.ps1
./testing/devnet/eip8321/Collect-Evidence.ps1
```

Use the Dora address printed by `Status.ps1`. Local ports are allocated dynamically and can change after a Docker restart.

Upstream Dora cannot decode the added EIP-8321 block body. For the feature-fork demonstration use the included read-only live view, which reads both beacon nodes directly:

```powershell
./testing/devnet/eip8321/Start-Dashboard.ps1
```

Open the printed URL. It shows node agreement, execution validation, finality, registration counts and hash reveals. Selecting a block displays its actual RANDAO fields. Python 3 is needed to serve the page on loopback; use `-PythonPath` if Python is outside PATH. `Stop-Dashboard.ps1` stops only that recorded server process.

Show these observations in order:

1. Before slot 32 the network uses Heze. Slot 32 begins epoch 4 and activates EIP-8321.
2. Validators generate and persist private hash-chain seeds, sign registrations and submit them to the beacon nodes. Blocks contain `randao_commitment_registrations`.
3. Registrations remain pending for the configured activation delay. A zero commitment continues to use the BLS RANDAO path.
4. After activation, proposing validators include a nonzero `hash_chain_reveal`; `randao_reveal` is the BLS infinity signature. The beacon node validates the hash preimage against its stored commitment.
5. Both nodes agree on a canonical block and finalization advances beyond the fork epoch. Check `execution_optimistic: false` too.

`Collect-Evidence.ps1` saves raw block JSON, fork and finality responses and a summary outside the repository. It fails if registrations, nonzero reveals, execution validation or post-fork finality are missing. Missed slots are recorded rather than counted as blocks. This is a live network check, separate from unit tests or consensus reference vectors.

For an independent hash check, install the small Python dependency and pass the evidence directory printed above:

```powershell
python -m pip install -r ./testing/devnet/eip8321/requirements.txt
python ./testing/devnet/eip8321/verify_hash_chain.py C:/path/to/evidence
```

This checks `BLAKE3("HASH_CHAIN_RANDAO" || reveal) == commitment`, follows each proposer's chain across blocks and enforces the registration delay. It uses the Python BLAKE3 binding independently of Prysm's Go implementation. It does not need private seeds or validator keys.

## Stop and repeat

```powershell
./testing/devnet/eip8321/Stop.ps1
```

This stops only the named enclave and retains its data. To demonstrate a fresh transition, use a new enclave name, or explicitly remove only the disposable demo enclave before starting again:

```powershell
kurtosis enclave rm -f bct-eip8321
./testing/devnet/eip8321/Start.ps1
```

Do not use a global Docker prune or stop unrelated containers. The validator database contains the private hash-chain seeds; deleting it while retaining an active network loses the ability to reveal registered commitments.

## Scope

See [VALIDATION.md](VALIDATION.md) for the recorded live-network observations, binary hashes and tested package scope.

The current implementation uses REST for Heze duties and gRPC for EIP-specific proposal, envelope and registration calls. Reading EIP blocks through REST includes the added fields; this does not implement every EIP REST publication or debug-state endpoint. Dora is an upstream explorer and its EIP body support must be checked separately from beacon-node correctness.

The generator emits newer `SLOT_DURATION_MS_HEZE` and `MIN_BLOB_DATA_RETENTION_MS` fields that this team baseline reports as unrecognized. This demonstration keeps the same 6000 ms slot duration across Heze; it does not validate a dynamic slot-duration change or historical blob-retention behavior.

This is a Prysm-only minimal-preset development network. Successful finalization does not establish interoperability with another client or full Ethereum mainnet readiness.
