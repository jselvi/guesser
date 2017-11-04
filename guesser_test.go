// Copyright 2017 Jose Selvi <jselvi{at}pentester.es>
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

package main

import (
    "testing"
)

const (
    exeRight = "sh guesser_test.sh"
    exeWrong = "cat guesser_test.sh"
    exeFail  = "doesntexist"
)

func Test_run(t *testing.T) {
    tests := []struct {
        name, cmd, param string
        want             int
        wantErr          bool
    }{
        {"check doesn't exist", exeFail, "xxx", -1, true},
        {"result not integer", exeWrong, "xxx", -1, true},
        {"check 123", exeRight, "123", 0, false},
        {"check eef", exeRight, "eef", 0, false},
        {"check 789", exeRight, "789", 1, false},
        {"check hjk", exeRight, "hjk", 1, false},
        {"check faa", exeRight, "faa", 0, false},
        {"check 234", exeRight, "234", 0, false},
        {"check zxc", exeRight, "zxc", 1, false},
        {"check caf", exeRight, "caf", 0, false},
        {"check xyz", exeRight, "xyz", 1, false},
        {"check 4ca", exeRight, "4ca", 0, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := run(tt.cmd, tt.param)
            if (err != nil) != tt.wantErr {
                t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("run() = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_score(t *testing.T) {
    tests := []struct {
        name, cmd, param string
        repeat, want     int
        wantErr          bool
    }{
        {"check 123", exeRight, "123", 5, 0, false},
        {"check 789", exeRight, "789", 5, 1, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := score(tt.cmd, tt.param, tt.repeat)
            if (err != nil) != tt.wantErr {
                t.Errorf("score() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("score() = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_sample(t *testing.T) {
    tests := []struct {
        name    string
        keys    []string
        want    string
        wantErr bool
    }{
        {"Empty map", []string{}, "", true},
        {"map add long", []string{"long"}, "long", false},
        {"map add srt", []string{"long", "srt"}, "long", false},
        {"map add longest", []string{"longest", "long", "srt"}, "longest", false},
    }

    for _, tt := range tests {
        var m = make(map[string]string)
        for _, key := range tt.keys {
            m[key] = "x"
        }
        t.Run(tt.name, func(t *testing.T) {
            got, err := sample(m)
            if (err != nil) != tt.wantErr {
                t.Errorf("sample() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("sample() = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_isAlreadyResult(t *testing.T) {
    tests := []struct {
        name string
        m map[string]bool
        s string
        want bool
    }{
        {"empty map", map[string]bool{},                            "whatever", false},
        {"found",     map[string]bool{"test": true, "found": true}, "oun",      true },
        {"not found", map[string]bool{"test": true, "lost": true},  "oun",      false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := isAlreadyResult(tt.m, tt.s); got != tt.want {
                t.Errorf("isAlreadyResult() = %v, want %v", got, tt.want)
            }
        })
    }
}
