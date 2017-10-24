// Copyright 2017 Jose Selvi <jselvi{at}pentester.es>
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

package main

import (
   "fmt"
   "sync"
   "io"
   "flag"
   "os/exec"
   "strconv"
   "strings"
   "time"
   "runtime"
   "errors"
)

// Dirty trick to run Cmd with unknown amount of params
func run(cmd string, param string) (int, error) {
    // Split Cmd
    v := strings.Split(cmd," ")
    var guess *exec.Cmd
    switch len(v) {
    case 1:
        guess = exec.Command( v[0] )
    case 2:
        guess = exec.Command( v[0], v[1] )
    case 3:
        guess = exec.Command( v[0], v[1], v[2] )
    case 4:
        guess = exec.Command( v[0], v[1], v[2], v[3] )
    case 5:
        guess = exec.Command( v[0], v[1], v[2], v[3], v[4] )
    case 6:
        guess = exec.Command( v[0], v[1], v[2], v[3], v[4], v[5] )
    default:
        return -1, errors.New("sorry, we couldn't parse Cmd")
    }

    stdin , _ := guess.StdinPipe()
    io.WriteString(stdin, param+"\n")
    out, err   := guess.Output()
    if err != nil {
        return -1, err
    }

    score, err := strconv.Atoi( strings.Split(string(out),"\n")[0] )
    if err != nil {
        return -1, err
    }

    return score, nil
}

// Gets score if "repeat" tries get the same result
func score(cmd string, param string, repeat int) (int, error) {
    res, _ := run(cmd, param)
    for i := 0; i < repeat-1; i++ {
        newres, _ := run(cmd, param)
        if res != newres {
            return -1, errors.New("Site seems to be unestable")
        }
    }
    return res, nil
}

// Gets longest key (more close to get a result)
func sample(m map[string]string) (string, error) {
    var l int = 0
    var key string
    for k, _ := range m {
        if len(k) > l {
            key = k
            l = len(k)
        }
    }
    if l > 0 {
        return key, nil
    } else {
        return "", errors.New("Empty Set")
    }
}

// Is "s" substring of any result from "m"?
func already_result(m map[string]bool, s string) bool {
    for k, _ := range m {
        if strings.Contains(k, s) {
            return true
        }
    }
    return false
}

// Main func
func main() {

    var cmd string
    var right string
    var wrong string
    var charset string
    var threads int
    var delay int

    // Params parsing
    flag.StringVar(&cmd, "cmd", "sh curl.sh", "command to run, parameter sent via stdin")
    flag.StringVar(&right, "right", " ", "term that makes cmd to give a right response")
    flag.StringVar(&wrong, "wrong", "^", "term that makes cmd to give a wrong response")
    flag.StringVar(&charset, "charset", "0123456789abcdef", "charset we use for guessing")
    flag.IntVar(&threads, "threads", 10, "amount of threads to use")
    flag.IntVar(&delay, "delay", 0, "delay between connections")
    flag.Parse()

    // Check stability
    score_right, err1 := score(cmd, right, 5)
    _          , err2 := score(cmd, wrong, 5)
    if (err1 != nil) || (err2 != nil) {
        fmt.Println("Unestable")
    }

    // Prepare a Set for substrings and a Set for results
    var pending = make(map[string]string)
    var tmp = make(map[string]bool)
    var res = make(map[string]bool)
    var mtx sync.Mutex
    pending[""] = "->"

    // While no pending strings to test, go for it
    for ( len(pending) > 0 ) {
        // Get a key
        key, _ := sample(pending)
        dir := pending[key]
        delete(pending, key)

        // If key is substring from a previous result, continue
        if len(key) > 1 && already_result(res, key) {
            continue
        }

        // Prepare Wait Group
        var wg sync.WaitGroup
        wg.Add( len(charset) )

        // Goroutines guessing
        for _, r := range charset {

            // Wait until we have available threads
            for runtime.NumGoroutine() >= threads+1 {
                time.Sleep(100 * time.Millisecond)
            }

            c := string(r)
            go func(pending map[string]string, cmd string, key string, dir string, c string, right int, res map[string]bool) {
                // Call done when gorouting ends 
                defer wg.Done()

                // Get term to test
                var term string
                if dir == "->" {
                    term = key+c
                } else {
                    term = c+key
                }

                // Calculate score
                score, _ := run(cmd, term)

                // Save results for next iteration
                if score == right {
                    mtx.Lock()
                    pending[term] = dir
                    mtx.Unlock()
                } else {
                    mtx.Lock()
                    tmp[term] = true
                    mtx.Unlock()
                }
            } (pending, cmd, key, dir, c, score_right, res)
        }

        // Wait for goroutines to finish
        wg.Wait()

        // If all chars were errors, we reached the start/end of a string 
        if len(tmp) == len(charset) {
            if dir == "->" {
                pending[key] = "<-"
            } else {
                res[key] = true
                fmt.Printf("\r%s\n", key)
            }
        } else {
            fmt.Printf("\r%s", key)
        }
        // Clean temporal map
        tmp = make(map[string]bool)
    }

    // Clean the last try
    fmt.Printf("\r                                                    \r")
}