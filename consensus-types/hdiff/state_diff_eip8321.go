package hdiff

import (
	"encoding/binary"
	"fmt"

	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
)

func (s *stateDiff) serializeEip8321Fields(data []byte) []byte {
	data = binary.LittleEndian.AppendUint64(data, uint64(len(s.randaoCommitments)))
	for _, c := range s.randaoCommitments {
		if len(c) != 32 {
			return nil
		}
		data = append(data, c...)
	}
	data = binary.LittleEndian.AppendUint64(data, uint64(len(s.pendingRandaoCommitments)))
	for _, p := range s.pendingRandaoCommitments {
		if p == nil || len(p.Commitment) != 32 {
			return nil
		}
		data = binary.LittleEndian.AppendUint64(data, uint64(p.ValidatorIndex))
		data = append(data, p.Commitment...)
		data = binary.LittleEndian.AppendUint64(data, uint64(p.ActivationEpoch))
	}
	return data
}

func (s *stateDiff) readEip8321Fields(data *[]byte) error {
	if len(*data) < 8 {
		return fmt.Errorf("truncated RANDAO commitments count")
	}
	count := binary.LittleEndian.Uint64((*data)[:8])
	*data = (*data)[8:]
	if count > uint64(len(*data)/32) {
		return fmt.Errorf("truncated RANDAO commitments")
	}
	s.randaoCommitments = make([][]byte, count)
	for i := range s.randaoCommitments {
		s.randaoCommitments[i] = append([]byte(nil), (*data)[:32]...)
		*data = (*data)[32:]
	}
	if len(*data) < 8 {
		return fmt.Errorf("truncated pending RANDAO commitments count")
	}
	count = binary.LittleEndian.Uint64((*data)[:8])
	*data = (*data)[8:]
	if count > uint64(len(*data)/48) {
		return fmt.Errorf("truncated pending RANDAO commitments")
	}
	s.pendingRandaoCommitments = make([]*ethpb.PendingRandaoCommitment, count)
	for i := range s.pendingRandaoCommitments {
		s.pendingRandaoCommitments[i] = &ethpb.PendingRandaoCommitment{
			ValidatorIndex:  primitives.ValidatorIndex(binary.LittleEndian.Uint64((*data)[:8])),
			Commitment:      append([]byte(nil), (*data)[8:40]...),
			ActivationEpoch: primitives.Epoch(binary.LittleEndian.Uint64((*data)[40:48])),
		}
		*data = (*data)[48:]
	}
	return nil
}
