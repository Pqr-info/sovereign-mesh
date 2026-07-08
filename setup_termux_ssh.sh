#!/bin/bash

# --- AESTHETIC CONSTANTS ---
CYAN="\033[96m"
GREEN="\033[92m"
GOLD="\033[93m"
RESET="\033[0m"
BOLD="\033[1m"

echo -e "${GOLD}======================================================================${RESET}"
echo -e "${BOLD}${CYAN}          SOVEREIGN MESH - TERMUX SSH CONFIGURATION TOOL${RESET}"
echo -e "${GOLD}======================================================================${RESET}"

SSH_DIR="$HOME/.ssh"
CONFIG_FILE="$SSH_DIR/config"

# Create .ssh directory if not exists
if [ ! -d "$SSH_DIR" ]; then
    mkdir -p "$SSH_DIR"
    chmod 700 "$SSH_DIR"
fi

# Define the SSH config block
HOST_ENTRY="Host antigravity-server
    HostName division-jaguar-renew-casio.trycloudflare.com
    User aellok
    ProxyCommand cloudflared access ssh --hostname %h
    IdentityFile ~/.ssh/id_ed25519"

# Check if the block already exists and remove/update it
if [ -f "$CONFIG_FILE" ]; then
    # Backup existing config
    cp "$CONFIG_FILE" "${CONFIG_FILE}.bak"
    # Remove existing host block for antigravity-server if it exists
    # Use python to perform regex/block removal reliably without depending on platform sed differences
    python3 -c "
import os
path = \"$CONFIG_FILE\"
with open(path, \"r\") as f:
    lines = f.read().splitlines()
new_lines = []
skip = False
for line in lines:
    if line.strip().startswith(\"Host antigravity-server\"):
        skip = True
        continue
    if skip:
        # Stop skipping once we hit another Host block
        if line.strip().startswith(\"Host \"):
            skip = False
        else:
            continue
    new_lines.append(line)
with open(path, \"w\") as f:
    f.write(\"\n\".join(new_lines) + \"\n\")
"
fi

# Append the new entry
echo -e "$HOST_ENTRY\n" >> "$CONFIG_FILE"
chmod 600 "$CONFIG_FILE"

echo -e "${GREEN}SUCCESS:${RESET} SSH configuration has been written to ${BOLD}$CONFIG_FILE${RESET}."
echo -e "You can now connect to the GCP instance from Termux by running:"
echo -e "  ${BOLD}ssh antigravity-server${RESET}"
echo -e "${GOLD}======================================================================${RESET}"
