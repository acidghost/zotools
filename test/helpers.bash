COVERAGE_PATH=${COVERAGE_PATH:-$(pwd)/.coverage}
ZOTOOLS_BIN=${ZOTOOLS_BIN:-$(pwd)/build/zotools}
STORAGE_SRC="$(pwd)/test/assets/storage.json"
STORAGE="$(pwd)/test/assets/storage.tmp.json"
STORAGE_MIME_UNKNOWN="$(pwd)/test/assets/storage_search_mime_unknown.json"
STORAGE_MIME_BROKEN="$(pwd)/test/assets/storage_search_mime_broken.json"
STORAGE_SINGLE_RES="$(pwd)/test/assets/storage_search_single_result.json"
STORAGE_EMPTY="$(pwd)/test/assets/storage_empty.json"

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

setup() {
    cp "$STORAGE_SRC" "$STORAGE"
}

teardown() {
    rm "$STORAGE"
}

cp_storage() {
    case "$1" in
        mime_unknown)
            cp "$STORAGE_MIME_UNKNOWN" "$STORAGE"
            ;;
        mime_broken)
            cp "$STORAGE_MIME_BROKEN" "$STORAGE"
            ;;
        single_result)
            cp "$STORAGE_SINGLE_RES" "$STORAGE"
            ;;
        empty)
            cp "$STORAGE_EMPTY" "$STORAGE"
            ;;
    esac
}

storage_contents() {
    cat "$STORAGE"
}
