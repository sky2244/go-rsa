package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

const TESTTARGET = 11 * 13
const DEBUG = false

type Rsa struct {
	Target  int64
	Answer1 int64
	Answer2 int64
	Indent  int64
}

func (r *Rsa) Is() bool {
	if r.Answer1 == r.Target || r.Answer2 == r.Target {
		return false
	}
	return r.Answer1*r.Answer2 == r.Target
}

func (r *Rsa) IsEnd() bool {
	return r.Answer1*r.Answer2 >= r.Target
}

func (r *Rsa) String() string {
	return fmt.Sprintf("Target: %d, Anser1: %d, Answer2: %d", r.Target, r.Answer1, r.Answer2)
}

func loop(ctx context.Context, wg *sync.WaitGroup, ch chan Rsa, cancel chan *Rsa, target Rsa) error {
	defer wg.Done()
	var i, u int64
	target_mod := target.Target % (target.Indent * 10)
	slog.DebugContext(ctx, fmt.Sprintf("Start: %#v, mod:%d", target, target_mod))
	if target.Is() {
		cancel <- &target
		return nil
	} else if target.IsEnd() {
		return nil
	}
	for i = 0; i < 10; i++ {
		for u = 0; u < 10; u++ {
			x := i*target.Indent + target.Answer1
			y := u*target.Indent + target.Answer2
			if (x*y)%(target.Indent*10) == target_mod {
				slog.DebugContext(ctx, fmt.Sprintf("add %#v  %d, %d", target, x, y))
				// add
				ch <- Rsa{
					Target:  target.Target,
					Answer1: x,
					Answer2: y,
					Indent:  target.Indent * 10,
				}
			}
		}
	}
	return nil
}

func main_loop(ctx context.Context, target int64) {
	var wg sync.WaitGroup
	loop_ctx, cancel := context.WithCancel(ctx)

	ch := make(chan Rsa)
	cancel_ch := make(chan *Rsa)
	target_rsa := Rsa{
		Target: target,
		Indent: 1,
	}

	go func() {
		for {
			select {
			case current := <-ch:
				wg.Add(1)
				go loop(loop_ctx, &wg, ch, cancel_ch, current)
			case res := <-cancel_ch:
				slog.InfoContext(ctx, fmt.Sprintf("Find! %s\n", res))
				cancel()
			}
		}
	}()

	wg.Add(1)
	loop(loop_ctx, &wg, ch, cancel_ch, target_rsa)
	go func() {
		wg.Wait()
		slog.InfoContext(ctx, fmt.Sprintf("Answer not found. Target: %d\n", target))
		cancel()
	}()

	<-loop_ctx.Done()
	slog.InfoContext(ctx, "finish")

}

func main() {
	ctx := context.Background()
	loglevel := slog.LevelInfo
	if DEBUG {
		loglevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: loglevel,
	}))
	slog.SetDefault(logger)
	slog.InfoContext(ctx, fmt.Sprintf("Target: %d", TESTTARGET))
	main_loop(ctx, TESTTARGET)
}
