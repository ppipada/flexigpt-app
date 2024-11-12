#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -f <flatpak_file_path> -d <analysis_dir_path> [-t <threshold>]"
    echo "  -f    Path to the flatpak file"
    echo "  -d    Path to the analysis directory"
    echo "  -t    Optional threshold (default: 500K)"
    exit 1
}

# Default threshold
threshold="500K"

# Parse command line arguments
while getopts ":f:d:t:" opt; do
    case "${opt}" in
        f)
            flatpak_file_path=$(realpath "${OPTARG}")
            ;;
        d)
            analysis_dir_path=$(realpath "${OPTARG}")
            ;;
        t)
            threshold=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done

# Check if mandatory arguments are provided
if [ -z "${flatpak_file_path}" ] || [ -z "${analysis_dir_path}" ]; then
    usage
fi

# Create analysis directory if it doesn't exist
echo "Creating analysis directory at ${analysis_dir_path}..."
mkdir -p "${analysis_dir_path}"

# Change to the analysis directory
echo "Changing to analysis directory..."
cd "${analysis_dir_path}" || exit

# Initialize an OSTree repository
echo "Initializing OSTree repository..."
ostree --repo=repo init --mode=archive

# Import the flatpak bundle into the repository
echo "Importing flatpak bundle into the repository..."
flatpak build-import-bundle repo "${flatpak_file_path}"

# List the refs in the repository and capture the output
echo "Listing refs in the repository..."
ref=$(ostree --repo=repo refs | head -n 1)
echo "Using ref: ${ref}"

# Checkout the application files using the ref
echo "Checking out application files..."
ostree --repo=repo checkout -U "${ref}" app_files

# Change to the app_files directory
echo "Changing to app_files directory..."
cd app_files || exit

# Perform the disk usage analysis with the specified threshold
echo "Performing disk usage analysis with threshold ${threshold}..."
du -ch --all -t "${threshold}" *

echo "Analysis complete."
