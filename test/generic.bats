#!/usr/bin/env bats -t

load helpers

@test "Version" {
    run_zotools -V
    [ "$status" -eq 0 ]
}

@test "No command" {
    run_zotools
    [ "$status" -eq 1 ]
}

@test "Unknown command" {
    run_zotools unknowncommand
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Command 'unknowncommand' not recognized" ]]
}

@test "Help command" {
    run_zotools help
    [ "$status" -eq 0 ]
    local pat='Usage: .+ \[OPTIONS\] command'
    [[ "$output" =~ $pat ]]
}

@test "Config invalid path" {
    local c=hopefullyanonexistingpath.json
    CONFIG=$c run_zotools search fuzz
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Failed to open config file" ]]
    [[ "$output" =~ "$c" ]]
}

@test "Config invalid values" {
    cp_config empty
    run_zotools search fuzz
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Wrong config values" ]]
    [[ "$output" =~ "key is empty" ]]
    [[ "$output" =~ "storage is empty" ]]
    [[ "$output" =~ "zotero is empty" ]]
}

@test "Search-act integration" {
    run_zotools search aflnet
    [ "$status" -eq 0 ]
    [[ "${lines[@]}" =~ "XZU8ER4Q" ]]

    run_zotools act echo
    [ "$status" -eq 0 ]
    [[ "${lines[1]}" =~ "XZU8ER4Q" ]]
}
