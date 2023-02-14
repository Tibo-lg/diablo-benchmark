#!/bin/bash

# This script uses ganache-cli for testing purposes

ganache-cli -m "nice charge tank ivory warfare spin deposit ecology beauty unusual comic melt" \
  -h "0.0.0.0" \
  -e 10000000000000000 \
  --wallet.accountKeysPath "accounts" \
  -a 15
  # -b 5
  # --verbose
