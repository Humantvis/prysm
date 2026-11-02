"""Independently check canonical block evidence using the Python BLAKE3 binding."""

import argparse
import json
from pathlib import Path

from blake3 import blake3


def verify(directory: Path, slots_per_epoch: int, delay: int) -> dict:
    commitments = {}
    activation_epochs = {}
    checks = []
    files = sorted(directory.glob("block-*.json"), key=lambda p: int(p.stem.split("-")[1]))
    if not files:
        raise ValueError("No block evidence found")
    for path in files:
        response = json.loads(path.read_text(encoding="utf-8-sig"))
        block = response["data"]["message"]
        slot = int(block["slot"])
        proposer = int(block["proposer_index"])
        body = block["body"]
        encoded = body.get("hash_chain_reveal")
        if encoded and bytes.fromhex(encoded.removeprefix("0x")) != bytes(32):
            reveal = bytes.fromhex(encoded.removeprefix("0x"))
            if len(reveal) != 32:
                raise ValueError(f"Slot {slot}: invalid reveal length")
            if proposer not in commitments:
                raise ValueError(f"Slot {slot}: missing registration for proposer {proposer}")
            if slot // slots_per_epoch < activation_epochs[proposer]:
                raise ValueError(f"Slot {slot}: revealed before registration activation")
            digest = blake3(b"HASH_CHAIN_RANDAO" + reveal).digest()
            if digest != commitments[proposer]:
                raise ValueError(f"Slot {slot}: BLAKE3(DST || reveal) differs from stored commitment")
            if body["randao_reveal"].lower() != "0xc0" + "00" * 95:
                raise ValueError(f"Slot {slot}: BLS infinity signature missing")
            checks.append({"slot": slot, "proposer_index": proposer, "commitment": "0x" + digest.hex(), "reveal": encoded})
            commitments[proposer] = reveal
        # RANDAO is processed before new registrations in the same block.
        for signed in body.get("randao_commitment_registrations", []):
            message = signed["message"]
            index = int(message["validator_index"])
            if index in commitments:
                raise ValueError(f"Duplicate registration for validator {index}")
            commitment = bytes.fromhex(message["commitment"].removeprefix("0x"))
            if len(commitment) != 32 or commitment == bytes(32):
                raise ValueError(f"Invalid commitment for validator {index}")
            commitments[index] = commitment
            activation_epochs[index] = slot // slots_per_epoch + delay
    if not checks:
        raise ValueError("No nonzero hash-chain reveals found")
    return {"registered_validators": len(commitments), "verified_hash_links": len(checks), "checks": checks}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("evidence", type=Path)
    parser.add_argument("--slots-per-epoch", type=int, default=8)
    parser.add_argument("--registration-delay", type=int, default=3)
    args = parser.parse_args()
    result = verify(args.evidence, args.slots_per_epoch, args.registration_delay)
    (args.evidence / "hash-verification.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"Verified {result['verified_hash_links']} BLAKE3 hash links; {result['registered_validators']} registered validators")
