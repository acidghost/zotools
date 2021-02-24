#!/usr/bin/env bats -t

load helpers

@test "Simple search" {
    run_zotools search something
    [ "$status" -eq 0 ]
}
