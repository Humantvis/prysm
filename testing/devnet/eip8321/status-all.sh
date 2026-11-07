#!/usr/bin/env bash
ENCLAVE=${1:-interop}
printf "%-22s %-6s %-20s %-12s %s\n" node slot root fork finalized
for svc in $(kurtosis enclave inspect $ENCLAVE 2>/dev/null | grep -o 'cl-[0-9]*-[a-z]*-[a-z]*' | sort -u); do
  url=$(kurtosis port print $ENCLAVE $svc http 2>/dev/null)
  head=$(curl -s $url/eth/v1/beacon/headers/head)
  slot=$(echo "$head" | jq -r .data.header.message.slot)
  root=$(echo "$head" | jq -r '.data.root[0:18]')
  fork=$(curl -s $url/eth/v1/beacon/states/head/fork | jq -r .data.current_version)
  fin=$(curl -s $url/eth/v1/beacon/states/head/finality_checkpoints | jq -r .data.finalized.epoch)
  printf "%-22s %-6s %-20s %-12s %s\n" $svc $slot "$root…" $fork $fin
done
