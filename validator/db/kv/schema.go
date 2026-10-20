package kv

var (
	// Genesis information bucket key.
	genesisInfoBucket = []byte("genesis-info-bucket")

	// Validator slashing protection from double proposals.
	historicProposalsBucket            = []byte("proposal-history-bucket-interchange")
	deprecatedAttestationHistoryBucket = []byte("attestation-history-bucket-interchange")

	// Buckets for lowest signed source and target epoch for individual validator.
	lowestSignedSourceBucket = []byte("lowest-signed-source-bucket")
	lowestSignedTargetBucket = []byte("lowest-signed-target-bucket")

	// Lowest and highest signed proposals.
	lowestSignedProposalsBucket  = []byte("lowest-signed-proposals-bucket")
	highestSignedProposalsBucket = []byte("highest-signed-proposals-bucket")

	// Slashable public keys bucket.
	slashablePublicKeysBucket = []byte("slashable-public-keys")

	// Genesis validators root bucket key.
	genesisValidatorsRootKey = []byte("genesis-val-root")

	// Optimized slashing protection buckets and keys.
	pubKeysBucket                 = []byte("pubkeys-bucket")
	attestationSigningRootsBucket = []byte("att-signing-roots-bucket")
	attestationSourceEpochsBucket = []byte("att-source-epochs-bucket")
	attestationTargetEpochsBucket = []byte("att-target-epochs-bucket")

	// Migrations
	migrationsBucket = []byte("migrations")

	// Graffiti
	graffitiBucket = []byte("graffiti")

	// Graffiti ordered index and hash keys
	graffitiOrderedIndexKey = []byte("graffiti-ordered-index")
	graffitiFileHashKey     = []byte("graffiti-file-hash")

	// ProposerSettings stores the encoded proposer settings file
	proposerSettingsBucket = []byte("proposer-settings-bucket")
	proposerSettingsKey    = []byte("proposer-settings")

	// Randao hash chain (EIP-8321): <pubkey> --> {seed || big-endian uint64 length}
	randaoHashChainBucket = []byte("randao-hash-chain-bucket")
)

// Attestations:
// -------------
// lowest-signed-source-bucket --> <pubkey> --> <epoch>
// lowest-signed-target-bucket --> <pubkey> --> <epoch>
//
// pubkeys-bucket --> <pubkey> --> att-signing-roots-bucket --> <target epoch> --> <signing root>
//                             |-> att-source-epochs-bucket --> <source epoch> --> []<target epoch>
//                             |-> att-target-epochs-bucket --> <target epoch> --> []<source epoch>

// Proposals:
// ----------
// proposal-history-bucket-interchange -> <pubkey> --> <slot> --> <signing root>

// EIP-8321 randao hash chain:
// ---------------------------
// randao-hash-chain-bucket --> <pubkey> --> <32-byte seed || 8-byte big-endian length>
// No chain, no position stored: the chain is recomputable from (seed, length)
// and the on-chain commitment is the position.
