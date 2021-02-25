#!/usr/bin/env bats -t

load helpers

@test "Act forget empty" {
    cp_storage empty
    run_zotools act -forget
    [ "$status" -eq 0 ]
    local pat='"Search":[[:space:]]*null'
    [[ "$(storage_contents)" =~ $pat ]]
}

@test "Act forget non-empty" {
    cp_storage single_result
    run_zotools act -forget
    [ "$status" -eq 0 ]
    local pat='"Search":[[:space:]]*null'
    [[ "$(storage_contents)" =~ $pat ]]
}

@test "Act no search" {
    cp_storage empty
    run_zotools act echo
    [ "$status" -eq 1 ]
    [[ "$output" =~ "No stored search" ]]
}

@test "Act index out of range" {
    cp_storage single_result
    run_zotools act -i=42 echo
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Index 42 is invalid" ]]
}

@test "Act unknown MIME" {
    cp_storage mime_unknown
    run_zotools act
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Unknown extension" ]]
}

@test "Act broken MIME" {
    cp_storage mime_broken
    run_zotools act
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Could not parse MIME type" ]]
}

@test "Act known MIME action" {
    cp_storage single_result
    ZOTOOLS_PDF=echo run_zotools act
    [ "$status" -eq 0 ]
    [[ "${lines[1]}" =~ "5D9UT6I4" ]]
}

@test "Act known MIME action broken command" {
    cp_storage single_result
    ZOTOOLS_PDF="echo -n 'Hello" run_zotools act
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Failed to parse ZOTOOLS_PDF" ]]
}

@test "Act unknown MIME action" {
    cp_storage single_result
    run_zotools act
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Command not found for MIME type" ]]
}

@test "Act command run error" {
    cp_storage single_result
    run_zotools act hopefullythisprogramisnotavailableanywhereweruntests
    [ "$status" -eq 1 ]
    [[ "${lines[0]}" =~ "5D9UT6I4" ]]
    [[ "$output" =~ "Failed to run action" ]]
}
