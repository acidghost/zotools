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

@test "Simple search" {
    run_zotools search learn
    [ "$status" -eq 0 ]
    [ "${lines[0]}" = "Loaded storage, version 771, 151 items" ]
    [[ "${lines[2]}" =~ '0)' ]]
    [[ "${lines[4]}" =~ '1)' ]]
    [[ "${lines[6]}" =~ '2)' ]]
    [[ "$output" =~ "Language-Agnostic Representation Learning" ]]
    [[ "$output" =~ "AZJXIBY6" ]]
    [[ "$output" =~ "Grammar-Based Fuzzing of REST" ]]
    [[ "$output" =~ "CQ3NJLC4" ]]
    [[ "$output" =~ "ES-Rank: evolution strategy learning to rank approach" ]]
    [[ "$output" =~ "ESB6N5MT" ]]
}

@test "Search uninfindable" {
    run_zotools search uninfindable
    [ "$status" -eq 1 ]
    [ "$output" = "Loaded storage, version 771, 151 items" ]
}

@test "Search authors" {
    run_zotools search -auth bohm
    [ "$status" -eq 0 ]
    [[ "$output" =~ "AFLNET: A Greybox Fuzzer for Network Protocols" ]]
    [[ "$output" =~ "STADS: Software Testing as Species Discovery" ]]
}

@test "Search abstract" {
    run_zotools search -abs performance
    [ "$status" -eq 0 ]
    [[ "$output" =~ "RetroWrite: Statically Instrumenting COTS Binaries for Fuzzing" ]]
}

@test "Search case-sensitive" {
    run_zotools search -s cots
    [ "$status" -eq 1 ]
    [[ ! "$output" =~ "RetroWrite: Statically Instrumenting COTS Binaries for Fuzzing" ]]
}

@test "Search invalid par" {
    run_zotools search -j=1337 aflnet
    [ "$status" -eq 1 ]
    [[ ! "$output" =~ "AFLNET" ]]
}

@test "Simple act" {
    run_zotools search aflnet
    [ "$status" -eq 0 ]
    [[ "${lines[@]}" =~ "XZU8ER4Q" ]]

    run_zotools act echo
    [ "$status" -eq 0 ]
    [[ "${lines[1]}" =~ "XZU8ER4Q" ]]
}

@test "Act no command" {
    run_zotools act
    [ "$status" -eq 1 ]
}

@test "Act index out of range" {
    run_zotools search aflnet
    [ "$status" -eq 0 ]
    [[ "${lines[@]}" =~ "XZU8ER4Q" ]]

    run_zotools act -i=42 echo
    [ "$status" -eq 1 ]
    [[ "$output" =~ "42" ]]
}
