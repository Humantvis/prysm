package structs

import (
	"encoding/json"
	"fmt"
	"strconv"

	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type RandaoCommitmentRegistration struct {
	ValidatorIndex string `json:"validator_index"`
	Commitment     string `json:"commitment"`
}

type SignedRandaoCommitmentRegistration struct {
	Message   *RandaoCommitmentRegistration `json:"message"`
	Signature string                        `json:"signature"`
}

type BeaconBlockBodyEip8321 struct {
	BeaconBlockBodyHeze
	HashChainReveal               string                                `json:"hash_chain_reveal"`
	RandaoCommitmentRegistrations []*SignedRandaoCommitmentRegistration `json:"randao_commitment_registrations"`
}

type BeaconBlockEip8321 struct {
	Slot          string                  `json:"slot"`
	ProposerIndex string                  `json:"proposer_index"`
	ParentRoot    string                  `json:"parent_root"`
	StateRoot     string                  `json:"state_root"`
	Body          *BeaconBlockBodyEip8321 `json:"body"`
}

type SignedBeaconBlockEip8321 struct {
	Message   *BeaconBlockEip8321 `json:"message"`
	Signature string              `json:"signature"`
}

var _ SignedMessageJsoner = &SignedBeaconBlockEip8321{}

func (s *SignedBeaconBlockEip8321) MessageRawJson() ([]byte, error) { return json.Marshal(s.Message) }
func (s *SignedBeaconBlockEip8321) SigString() string               { return s.Signature }

func SignedBeaconBlockEip8321FromConsensus(s *ethpb.SignedBeaconBlockEip8321) (*SignedBeaconBlockEip8321, error) {
	if s == nil || s.Block == nil || s.Block.Body == nil {
		return nil, fmt.Errorf("nil EIP-8321 block")
	}
	b := s.Block
	body := b.Body
	registrations := make([]*SignedRandaoCommitmentRegistration, len(body.RandaoCommitmentRegistrations))
	for i, r := range body.RandaoCommitmentRegistrations {
		if r == nil || r.Message == nil {
			return nil, fmt.Errorf("nil RANDAO registration at index %d", i)
		}
		registrations[i] = &SignedRandaoCommitmentRegistration{
			Message:   &RandaoCommitmentRegistration{ValidatorIndex: strconv.FormatUint(uint64(r.Message.ValidatorIndex), 10), Commitment: hexutil.Encode(r.Message.Commitment)},
			Signature: hexutil.Encode(r.Signature),
		}
	}
	attestations := make([]*PayloadAttestation, len(body.PayloadAttestations))
	for i, a := range body.PayloadAttestations {
		attestations[i] = PayloadAttestationFromConsensus(a)
	}
	return &SignedBeaconBlockEip8321{
		Signature: hexutil.Encode(s.Signature),
		Message: &BeaconBlockEip8321{
			Slot: strconv.FormatUint(uint64(b.Slot), 10), ProposerIndex: strconv.FormatUint(uint64(b.ProposerIndex), 10),
			ParentRoot: hexutil.Encode(b.ParentRoot), StateRoot: hexutil.Encode(b.StateRoot),
			Body: &BeaconBlockBodyEip8321{
				BeaconBlockBodyHeze: BeaconBlockBodyHeze{
					RandaoReveal: hexutil.Encode(body.RandaoReveal), Graffiti: hexutil.Encode(body.Graffiti),
					ProposerSlashings: ProposerSlashingsFromConsensus(body.ProposerSlashings),
					AttesterSlashings: AttesterSlashingsGloasFromConsensus(body.AttesterSlashings), Attestations: AttsGloasFromConsensus(body.Attestations),
					VoluntaryExits: SignedExitsFromConsensus(body.VoluntaryExits), SyncAggregate: SyncAggregateFromConsensus(body.SyncAggregate),
					BLSToExecutionChanges:     SignedBLSChangesFromConsensus(body.BlsToExecutionChanges),
					SignedExecutionPayloadBid: SignedExecutionPayloadBidHezeFromConsensus(body.SignedExecutionPayloadBid),
					PayloadAttestations:       attestations, ParentExecutionRequests: ExecutionRequestsGloasFromConsensus(body.ParentExecutionRequests),
				},
				HashChainReveal: hexutil.Encode(body.HashChainReveal), RandaoCommitmentRegistrations: registrations,
			},
		},
	}, nil
}
