# Genesis generator that also writes the EIP-8321 fork into the CL config.yaml.
# Values come from ethereum_genesis_generator_params.extra_env in devnet-eip8321-minimal.yaml;
# the ENV defaults (fork unscheduled) only apply when extra_env omits them.
# 6.2.9 is the default of the pinned ethereum-package@a4f2204: bump both together.
FROM ethpandaops/ethereum-genesis-generator:6.2.9
ENV EIP8321_FORK_VERSION=0xe8321001 EIP8321_FORK_EPOCH=18446744073709551615
RUN printf '\n# EIP-8321 (Hash-Chain RANDAO)\nEIP8321_FORK_VERSION: $EIP8321_FORK_VERSION\nEIP8321_FORK_EPOCH: $EIP8321_FORK_EPOCH\n' >> /config/cl/config.yaml
