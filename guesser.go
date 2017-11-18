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

const (
    defaultCmd     = "sh curl.sh"
    defaultRight   = " "
    defaultWrong   = "^"
    defaultCharset = "0123456789abcdef"
    defaultInit    = ""
    defaultThreads = 10
    defaultDelay   = 0
)

// Dirty trick to run Cmd with unknown amount of params
func run(cmd string, param string) (int, error) {
    // Split Cmd
    v := strings.Split(cmd," ")
    guess := exec.Command( v[0], v[1:]... )

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
    var l int
    var key string
    for k := range m {
        if len(k) > l {
            key = k
            l = len(k)
        }
    }
    if l > 0 {
        return key, nil
    }
    return "", errors.New("Empty Set")
}

// Is "s" substring of any result from "m"?
func isAlreadyResult(m map[string]bool, s string) bool {
    for k := range m {
        if strings.Contains(k, s) {
            return true
        }
    }
    return false
}

// Main func
func main() {
    // Params parsing
    cmd     := flag.String("cmd"    , defaultCmd    , "command to run, parameter sent via stdin")
    right   := flag.String("right"  , defaultRight  , "term that makes cmd to give a right response")
    wrong   := flag.String("wrong"  , defaultWrong  , "term that makes cmd to give a wrong response")
    charset := flag.String("charset", defaultCharset, "charset we use for guessing")
    init    := flag.String("init"   , defaultInit   , "Initial search string")
    threads := flag.Int(   "threads", defaultThreads, "amount of threads to use")
    delay   := flag.Int(   "delay"  , defaultDelay  , "delay between connections")
    flag.Parse()

    // Call to the main func
    guessIt( cmd, right, wrong, charset, init, threads, delay, false)
}

// Gets arguments from map instead of command line (for testing purposes)
func guessItMap( param map[string]string ) map[string]bool {
    var cmd     = defaultCmd
    var right   = defaultRight
    var wrong   = defaultWrong
    var charset = defaultCharset
    var init    = defaultInit
    var threads = defaultThreads
    var delay   = defaultDelay
    var err error

    for name, value := range param {
        switch name {
        case "cmd":
            cmd = value
        case "right":
            right = value
        case "wrong":
            wrong = value
        case "charset":
            charset = value
        case "init":
            init = value
        case "threads":
            threads, err = strconv.Atoi(value)
            if err != nil {
                threads = defaultThreads
            }
        case "delay":
            delay, err = strconv.Atoi(value)
            if err != nil {
                delay = defaultDelay
            }
        }
    }

    return guessIt(&cmd, &right, &wrong, &charset, &init, &threads, &delay, true)
}

// Real core
func guessIt( cmd, right, wrong, charset, init *string, threads, delay *int, quiet bool) map[string]bool {
    // Check stability
    scoreRight, err1 := score(*cmd, *right, 5)
    _         , err2 := score(*cmd, *wrong, 5)
    if (err1 != nil) || (err2 != nil) {
        if !quiet {
            fmt.Println("Unstable")
        }
    }

    // Prepare a Set for substrings and a Set for results
    var pending = make(map[string]string)
    var tmp = make(map[string]bool)
    var res = make(map[string]bool)
    var mtx sync.Mutex
    pending[*init] = "->"

    // While no pending strings to test, go for it
    for ( len(pending) > 0 ) {
        // Get a key
        key, _ := sample(pending)
        dir := pending[key]
        delete(pending, key)

        // If key is substring from a previous result, continue
        if len(key) > len(*init)+1 && isAlreadyResult(res, key) {
            continue
        }

        // Prepare Wait Group
        var wg sync.WaitGroup
        wg.Add( len(*charset) )

        // Goroutines guessing
        for _, r := range *charset {

            // Wait until we have available threads
            for runtime.NumGoroutine() >= (*threads)+1 {
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
            } (pending, *cmd, key, dir, c, scoreRight, res)
        }

        // Wait for goroutines to finish
        wg.Wait()

        // If all chars were errors, we reached the start/end of a string 
        if len(tmp) == len(*charset) {
            if dir == "->" {
                pending[key] = "<-"
            } else {
                res[key] = true
                if !quiet {
                    fmt.Printf("\r%s\n", key)
                }
            }
        } else {
            if !quiet {
                fmt.Printf("\r%s", key)
            }
        }
        // Clean temporal map
        tmp = make(map[string]bool)
    }

    // Clean the last try
    if !quiet {
        fmt.Printf("\r                                                    \r")
    }

    return res
}
