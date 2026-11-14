package service_test

// Throughput benchmarks over loopback (TLS + SHA-256 + disk). They are skipped
// unless the inputs are given, e.g.:
//
//	LT_BENCH_FILE=D:ig.bin LT_BENCH_DIR=D:\many-files LT_BENCH_OUT=D:	mp //	  go test -run XXX -bench . -benchtime 3x ./internal/service/

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"lantransfer/internal/app"
	"lantransfer/internal/security"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
)

func benchSend(b *testing.B, src string, parallel int) {
	if src == "" {
		b.Skip("LT_BENCH_* not set")
	}
	var total int64
	filepath.Walk(src, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	identity, _ := security.NewIdentity("bench")
	root := os.Getenv("LT_BENCH_OUT")
	state := &app.State{ID: "r", Name: "R", Identity: identity, Pairing: security.NewPairing("111111"), History: transfer.LoadHistory("")}
	listener, _ := net.Listen("tcp4", "127.0.0.1:0")
	dirs := []string{}
	receiver := &transfer.Receiver{}
	server := service.NewServerWith(state, receiver)
	go server.ServeTLS(listener, "", "")
	defer server.Close()
	b.SetBytes(total)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dir, _ := os.MkdirTemp(root, "recv")
		dirs = append(dirs, dir)
		receiver.Directory = dir
		b.StartTimer()
		s, err := transfer.Pair(context.Background(), listener.Addr().String(), "111111", state.Info())
		if err != nil {
			b.Fatal(err)
		}
		plan, _ := transfer.PlanPath(src, s.Address, "")
		if err := s.Send(context.Background(), plan, "", parallel, nil); err != nil {
			b.Fatal(err)
		}
		s.Close()
		b.StopTimer()
		os.RemoveAll(dir)
		b.StartTimer()
	}
}

func BenchmarkLargeFile(b *testing.B)      { benchSend(b, os.Getenv("LT_BENCH_FILE"), 4) }
func BenchmarkManySmallFiles(b *testing.B) { benchSend(b, os.Getenv("LT_BENCH_DIR"), 4) }

