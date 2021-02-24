COVERAGE_PATH=${COVERAGE_PATH:-$(pwd)/.coverage}
ZOTOOLS_BIN=${ZOTOOLS_BIN:-$(pwd)/build/zotools}

VERSION=$(cat VERSION)
export VERSION

random_string() {
    local length=${1:-10}

    head /dev/urandom | tr -dc a-zA-Z0-9 | head -c"$length"
}

run_zotools() {
    local args=()
    if [[ -n "$COVERAGE" ]]; then
        args+=("-test.coverprofile=coverage.integration.$(random_string 20).txt")
        args+=("-test.outputdir=$COVERAGE_PATH" COVERAGE)
    fi
    args+=("-config=$(pwd)/test/assets/config.json")
    echo "[+] Running $ZOTOOLS_BIN" "${args[@]}" "$*"
    run "$ZOTOOLS_BIN" "${args[@]}" "$@"
    # shellcheck disable=SC2154
    if [ "$status" -ne 0 ]; then
        echo "$output"
    fi
}
