package randaocommitments

import (
	"maps"
	"sync"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	doublylinkedlist "github.com/OffchainLabs/prysm/v7/container/doubly-linked-list"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/pkg/errors"
)

// We recycle the pool's map to avoid the backing map growing without bound.
// The cycling operation is expensive because it copies all elements, so we
// only do it when the map is smaller than this upper bound.
const randaoCommitmentsPoolThreshold = 2000

var (
	randaoCommitmentInPoolTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "randao_commitment_registration_pool_total",
		Help: "The number of saved randao commitment registrations in the operation pool.",
	})
)

// PoolManager maintains pending and seen randao commitment registrations.
// This pool is used by proposers to insert registrations into new blocks.
type PoolManager interface {
	PendingRandaoCommitments() ([]*ethpb.SignedRandaoCommitmentRegistration, error)
	RandaoCommitmentsForInclusion(beaconState state.ReadOnlyBeaconState) ([]*ethpb.SignedRandaoCommitmentRegistration, error)
	InsertRandaoCommitment(msg *ethpb.SignedRandaoCommitmentRegistration)
	MarkIncluded(msg *ethpb.SignedRandaoCommitmentRegistration)
	ValidatorExists(idx primitives.ValidatorIndex) bool
}

// Pool is a concrete implementation of PoolManager.
type Pool struct {
	lock    sync.RWMutex
	pending doublylinkedlist.List[*ethpb.SignedRandaoCommitmentRegistration]
	m       map[primitives.ValidatorIndex]*doublylinkedlist.Node[*ethpb.SignedRandaoCommitmentRegistration]
}

// NewPool returns an initialized pool.
func NewPool() *Pool {
	return &Pool{
		pending: doublylinkedlist.List[*ethpb.SignedRandaoCommitmentRegistration]{},
		m:       make(map[primitives.ValidatorIndex]*doublylinkedlist.Node[*ethpb.SignedRandaoCommitmentRegistration]),
	}
}

// Copies the internal map and returns a new one.
func (p *Pool) cycleMap() {
	newMap := make(map[primitives.ValidatorIndex]*doublylinkedlist.Node[*ethpb.SignedRandaoCommitmentRegistration])
	maps.Copy(newMap, p.m)
	p.m = newMap
}

// PendingRandaoCommitments returns all objects from the pool.
func (p *Pool) PendingRandaoCommitments() ([]*ethpb.SignedRandaoCommitmentRegistration, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	result := make([]*ethpb.SignedRandaoCommitmentRegistration, p.pending.Len())
	node := p.pending.First()
	var err error
	for i := 0; node != nil; i++ {
		result[i], err = node.Value()
		if err != nil {
			return nil, err
		}
		node, err = node.Next()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// RandaoCommitmentsForInclusion returns objects that are ready for inclusion.
// It returns at most MAX_RANDAO_COMMITMENT_REGISTRATIONS entries and drops any
// entry that is no longer valid against the given state.
func (p *Pool) RandaoCommitmentsForInclusion(st state.ReadOnlyBeaconState) ([]*ethpb.SignedRandaoCommitmentRegistration, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	length := int(min(float64(params.BeaconConfig().MaxRandaoCommitmentRegistrations), float64(p.pending.Len())))
	result := make([]*ethpb.SignedRandaoCommitmentRegistration, 0, length)

	// Validators that already have a pending registration in the state are invalid for a block.
	pendingInState, err := st.PendingRandaoCommitments()
	if err != nil {
		return nil, err
	}
	pendingIdx := make([]uint64, len(pendingInState))
	for i, pc := range pendingInState {
		pendingIdx[i] = uint64(pc.ValidatorIndex)
	}

	node := p.pending.Last()
	for node != nil && len(result) < length {
		msg, err := node.Value()
		if err != nil {
			return nil, err
		}
		// Take the previous node first: removing this one may invalidate its links.
		prev, prevErr := node.Prev()

		if err := invalidReason(st, msg, pendingIdx); err != nil {
			logrus.WithError(err).WithField("validatorIndex", msg.Message.ValidatorIndex).
				Warning("removing invalid RandaoCommitmentRegistration from pool")
			// MarkIncluded removes the invalid message from the pool.
			p.lock.RUnlock()
			p.MarkIncluded(msg)
			p.lock.RLock()
		} else {
			result = append(result, msg)
		}
		if prevErr != nil {
			return nil, prevErr
		}
		node = prev
	}
	return result, nil
}

// invalidReason returns an error if the message can no longer go into a block.
// The signature was verified at gossip time, and the domain uses the genesis fork
// version, so it cannot go stale across forks and is not rechecked here.
func invalidReason(st state.ReadOnlyBeaconState, msg *ethpb.SignedRandaoCommitmentRegistration, pendingIdx []uint64) error {
	if msg == nil || msg.Message == nil || len(msg.Message.Commitment) != 32 {
		return errors.New("malformed registration")
	}
	var commitment [32]byte
	copy(commitment[:], msg.Message.Commitment)
	idx := msg.Message.ValidatorIndex
	numValidators := uint64(st.NumValidators())
	var stored [32]byte
	if uint64(idx) < numValidators {
		var err error
		stored, err = st.RandaoCommitmentAtIndex(idx)
		if err != nil {
			return err
		}
	}
	return eip8321.VerifyRandaoCommitmentRegistration(uint64(idx), numValidators, commitment, stored, pendingIdx)
}

// InsertRandaoCommitment inserts an object into the pool.
func (p *Pool) InsertRandaoCommitment(msg *ethpb.SignedRandaoCommitmentRegistration) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, exists := p.m[msg.Message.ValidatorIndex]; exists {
		return
	}

	p.pending.Append(doublylinkedlist.NewNode(msg))
	p.m[msg.Message.ValidatorIndex] = p.pending.Last()

	randaoCommitmentInPoolTotal.Inc()
}

// MarkIncluded is used when an object has been included in a beacon block. Every block seen by this
// node should call this method. This will remove the object from the pool.
func (p *Pool) MarkIncluded(msg *ethpb.SignedRandaoCommitmentRegistration) {
	p.lock.Lock()
	defer p.lock.Unlock()

	node := p.m[msg.Message.ValidatorIndex]
	if node == nil {
		return
	}

	delete(p.m, msg.Message.ValidatorIndex)
	p.pending.Remove(node)
	if p.numPending() == randaoCommitmentsPoolThreshold {
		p.cycleMap()
	}

	randaoCommitmentInPoolTotal.Dec()
}

// ValidatorExists checks if a registration exists in the pool for that validator.
func (p *Pool) ValidatorExists(idx primitives.ValidatorIndex) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.m[idx] != nil
}

// numPending returns the number of pending registrations in the pool.
func (p *Pool) numPending() int {
	return p.pending.Len()
}