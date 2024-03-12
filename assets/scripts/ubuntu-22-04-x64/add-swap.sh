#!/bin/bash

# Add swap space if not already added
SWAP_EXIST=$(sudo swapon --show)
SWAP_SIZE=2G
SWAP_FILE=/swapfile
if [ -z "$SWAP_EXIST" ]; then
    sudo fallocate -l $SWAP_SIZE $SWAP_FILE || sudo dd if=/dev/zero of=$SWAP_FILE bs=1024 count=$(echo $SWAP_SIZE | grep -o -E '[0-9]+')k
    sudo chmod 600 $SWAP_FILE
    sudo mkswap $SWAP_FILE
    sudo swapon $SWAP_FILE
fi
sudo swapon --show