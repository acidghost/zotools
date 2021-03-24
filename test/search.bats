#!/usr/bin/env bats -t

load helpers

@test "Simple search" {
    run_zotools search learn
    [ "$status" -eq 0 ]
    [ "${lines[0]}" = "Loaded storage, version 779, 153 items" ]
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
    [ "$output" = "Loaded storage, version 779, 153 items" ]
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
