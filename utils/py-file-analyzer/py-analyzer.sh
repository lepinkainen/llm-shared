#!/bin/bash
# Python File Analyzer - Wrapper script
# Usage: ./py-analyzer.sh [directory] [num_files] [top_functions] [include_tests]

set -e

# Default values
DIR=${1:-"."}
NUM_FILES=${2:-20}
TOP_FUNCS=${3:-5}
INCLUDE_TESTS=${4:-false}

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run the analyzer with the provided arguments
cd "$(dirname "$DIR")"  # Change to parent directory of target
ARGS=("-dir" "$(basename "$DIR")" "-n" "$NUM_FILES" "-topfuncs" "$TOP_FUNCS")

if [ "$INCLUDE_TESTS" = "true" ]; then
    ARGS+=("--include-tests")
fi

uv run python "$SCRIPT_DIR/main.py" "${ARGS[@]}"