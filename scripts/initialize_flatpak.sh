#!/bin/bash

echo "Starting Initializing Flatpak for apt based systems (Ubuntu, etc)..."
# Update package lists
echo "Updating package lists..."
sudo apt update

# Install Flatpak
echo "Installing Flatpak..."
sudo apt install -y flatpak

# Install the GNOME Software plugin for Flatpak
echo "Installing GNOME Software plugin for Flatpak..."
sudo apt install -y gnome-software-plugin-flatpak

# Add the Flathub repository
echo "Adding Flathub repository..."
flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo

# Notify user of completion
echo "Flatpak has been initialized successfully. You may need to restart your system for changes to take effect."
echo "Some useful commands:"
echo "  flatpak install --user FlexiGPT.flatpak"
echo "  flatpak list"
echo "  flatpak info io.github.flexigpt.client"
echo "  flatpak run io.github.flexigpt.client"
echo "  cat ~/.var/app/io.github.flexigpt.client/config/flexigpt/settings.json"
echo "  ls -la ~/.var/app/io.github.flexigpt.client/data/flexigpt/conversations"
echo "  ls -la ~/.var/app/io.github.flexigpt.client/data/flexigpt/logs"
