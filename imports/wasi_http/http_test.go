package wasi_http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

type handler struct {
	urls []string
}

func (h *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Write([]byte("Response"))

	h.urls = append(h.urls, req.URL.String())
}

func TestHttp(t *testing.T) {
	filePaths, _ := filepath.Glob("../../testdata/c/http/*.wasm")
	for _, file := range filePaths {
		fmt.Printf("%v\n", file)
	}
	if len(filePaths) == 0 {
		t.Log("nothing to test")
	}

	h := handler{}
	s := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: &h,
	}
	go s.ListenAndServe()
	defer s.Shutdown(context.TODO())

	expectedPaths := [][]string{
		[]string{"/get?some=arg&goes=here"},
	}

	for testIx, test := range filePaths {
		name := test
		for strings.HasPrefix(name, "../") {
			name = name[3:]
		}

		t.Run(name, func(t *testing.T) {
			bytecode, err := os.ReadFile(test)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			runtime := wazero.NewRuntime(ctx)
			defer runtime.Close(ctx)

			builder := imports.NewBuilder().
				WithName("http").
				WithArgs()
			var system wasi.System
			ctx, system, err = builder.Instantiate(ctx, runtime)
			if err != nil {
				t.Error("Failed to build WASI module: ", err)
			}
			defer system.Close(ctx)

			Instantiate(ctx, runtime)

			instance, err := runtime.Instantiate(ctx, bytecode)
			if err != nil {
				switch e := err.(type) {
				case *sys.ExitError:
					if exitCode := e.ExitCode(); exitCode != 0 {
						t.Error("exit code:", exitCode)
					}
				default:
					t.Error("instantiating wasm module instance:", err)
				}
			}
			if instance != nil {
				if err := instance.Close(ctx); err != nil {
					t.Error("closing wasm module instance:", err)
				}
			}
			ok := true
			if len(h.urls) != len(expectedPaths[testIx]) {
				ok = false
			} else {
				for ix := range h.urls {
					if h.urls[ix] != expectedPaths[testIx][ix] {
						ok = false
						break
					}
				}
			}
			if !ok {
				t.Errorf("Unexpected paths: %v vs %v", h.urls, expectedPaths[testIx])
			}
			h.urls = []string{}
		})
	}
}
