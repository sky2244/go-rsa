package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"sync"
)

const (
	LIMIT = 8
	DEBUG = true
)

var (
	TESTTARGET *big.Int
)

type Rsa struct {
	Target  *big.Int
	Answer1 *big.Int
	Answer2 *big.Int
	Indent  *big.Int
}

func (r *Rsa) Is() bool {
	if r.Answer1.Cmp(r.Target) == 0 || r.Answer2.Cmp(r.Target) == 0 {
		return false
	}
	var mul big.Int
	mul = *mul.Mul(r.Answer1, r.Answer2)
	return mul.Cmp(r.Target) == 0
}

func (r *Rsa) IsEnd() bool {
	var mul big.Int
	mul = *mul.Mul(r.Answer1, r.Answer2)
	return mul.Cmp(r.Target) >= 0
}

func (r Rsa) String() string {
	return fmt.Sprintf("Target: %s, Anser1: %s, Answer2: %s, Indent: %s", r.Target.String(), r.Answer1.String(), r.Answer2.String(), r.Indent.String())
}

func loop(ctx context.Context, wg *sync.WaitGroup, ch chan *Rsa, cancel chan *Rsa, target *Rsa) error {
	defer wg.Done()
	var i, u, mod_indent, target_mod big.Int
	mod_indent = *big.NewInt(0)
	target_mod = *big.NewInt(0)
	slog.DebugContext(ctx, fmt.Sprintf("Start: %s, mod:%s, %p", target.String(), target_mod.String(), target))
	mod_indent.Mul(target.Indent, big.NewInt(10))
	target_mod.Mod(target.Target, &mod_indent)

	if target.Is() {
		cancel <- target
		return nil
	} else if target.IsEnd() {
		return nil
	}
	loop_end := big.NewInt(10)
	loop_end = loop_end.Mul(loop_end, target.Indent)
	for i = *big.NewInt(0); i.Cmp(loop_end) < 0; i.Add(&i, target.Indent) {
		for u = *big.NewInt(0); u.Cmp(loop_end) < 0; u.Add(&u, target.Indent) {
			var x big.Int
			x.Add(target.Answer1, &i)
			var y, xy big.Int
			y.Add(target.Answer2, &u)

			xy.Mul(&x, &y)
			xy.Mod(&xy, &mod_indent)
			if xy.Cmp(&target_mod) == 0 {
				slog.DebugContext(ctx, fmt.Sprintf("add %s  x:%s, y:%s, xy:%s, i:%s,u:%s, %p", target.String(), x.String(), y.String(), xy.String(), i.String(), u.String(), target))
				// add
				var indent big.Int
				indent = *indent.Set(target.Indent)
				indent = *indent.Mul(&indent, big.NewInt(10))
				ch <- &Rsa{
					Target:  target.Target,
					Answer1: &x,
					Answer2: &y,
					Indent:  &indent,
				}
			}
		}
	}
	return nil
}

func main_loop(ctx context.Context, target *big.Int) {
	var wg sync.WaitGroup
	loop_ctx, cancel := context.WithCancel(ctx)

	ch := make(chan *Rsa, 1024)
	cancel_ch := make(chan *Rsa)
	target_rsa := Rsa{
		Target:  target,
		Answer1: big.NewInt(1),
		Answer2: big.NewInt(1),
		Indent:  big.NewInt(1),
	}

	slots := make(chan struct{}, LIMIT)

	l := NewHeap()
	go func() {
		for {
			current := l.Pop()

			wg.Add(1)
			slots <- struct{}{}
			go func() {
				loop(loop_ctx, &wg, ch, cancel_ch, current.(*Rsa))
				<-slots
			}()
		}
	}()

	go func() {
		for {
			select {
			case current := <-ch:
				l.Push(current)
			case res := <-cancel_ch:
				slog.InfoContext(ctx, fmt.Sprintf("Find! %s", res))
				cancel()
			}
		}
	}()

	wg.Add(1)
	loop(loop_ctx, &wg, ch, cancel_ch, &target_rsa)
	go func() {
		wg.Wait()
		slog.InfoContext(ctx, fmt.Sprintf("Answer not found. Target: %s", target.String()))
		// cancel()
	}()

	<-loop_ctx.Done()
	slog.InfoContext(ctx, "finish")

}

func test_set_target() {
	var p, q big.Int
	// sample
	p.SetString("11", 10)
	q.SetString("13", 10)
	// 10
	// p.SetString("1833206873", 10)
	// q.SetString("2506048141", 10)
	TESTTARGET = big.NewInt(0)
	TESTTARGET.Mul(&p, &q)
}

func main() {
	test_set_target()
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
