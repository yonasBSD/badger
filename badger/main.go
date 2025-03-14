/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"runtime"

	"github.com/dustin/go-humanize"
	"go.opentelemetry.io/contrib/zpages"

	"github.com/dgraph-io/badger/v4/badger/cmd"
	"github.com/dgraph-io/ristretto/v2/z"
)

func main() {
	go func() {
		for i := 8080; i < 9080; i++ {
			fmt.Printf("Listening for /debug HTTP requests at port: %d\n", i)
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", i), nil); err != nil {
				fmt.Println("Port busy. Trying another one...")
				continue

			}
		}
	}()
	http.DefaultServeMux.Handle("/z", zpages.NewTracezHandler(zpages.NewSpanProcessor()))
	runtime.SetBlockProfileRate(100)
	runtime.GOMAXPROCS(128)

	out := z.CallocNoRef(1, "Badger.Main")
	fmt.Printf("jemalloc enabled: %v\n", len(out) > 0)
	z.StatsPrint()
	z.Free(out)

	cmd.Execute()
	fmt.Printf("Num Allocated Bytes at program end: %s\n",
		humanize.IBytes(uint64(z.NumAllocBytes())))
	if z.NumAllocBytes() > 0 {
		fmt.Println(z.Leaks())
	}
}
