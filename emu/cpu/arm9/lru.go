package arm9

// block cache assigns and returns blocks to be used by jit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aabalke/gojit"
)

type BlockCache struct {
    Head, Tail *JitBlock
    Blocks []*JitBlock
    mu sync.Mutex
}

type JitBlock struct {
	Skip      bool
	f         func()
	initPc    uint32
	finalOp   uint32
	Length    uint32
	assembler *gojit.Assembler

    Prev, Next *JitBlock

    refs atomic.Int32
}

func InitBlockCache(capacity, page_size int) *BlockCache {

    bc := &BlockCache{
        Blocks: []*JitBlock{},
        Head: &JitBlock{},
        Tail: &JitBlock{},
    }

    for range capacity {
        bc.Blocks = append(bc.Blocks, &JitBlock{})
    }

    // head and tail are never directly used, just pivots for
    // head prev and tail next
    bc.Head.Next = bc.Tail
    bc.Tail.Prev = bc.Head

    for i := range capacity {

        bc.Blocks[i].Prev = bc.Head
        bc.Blocks[i].Next = bc.Head.Next

        bc.Head.Next.Prev = bc.Blocks[i]
        bc.Head.Next = bc.Blocks[i]

        asm, err := gojit.New(page_size)
        if err != nil {
            panic(err)
        }

        if asm == nil {
            panic("bad asm init")
        }

        bc.Blocks[i].assembler = asm
    }

    return bc
}

func (bc *BlockCache) String() string {

    // string, tail -> head pointers
    s := "B "
    for b := bc.Tail;; {
        if b != bc.Tail && b != bc.Head {
            s += fmt.Sprintf("%p ", b)
        }
        b = b.Prev
        if b.Prev == nil {
            break
        }
    }

    return s + "\n"
}

func (bc *BlockCache) BlockString() string {

    // string, need tail and head included to match String()

    s := fmt.Sprintf("B %p ", bc.Tail)
    for i := range len(bc.Blocks) {
        s += fmt.Sprintf("%p ", bc.Blocks[i])
    }

    s += fmt.Sprintf("%p\n", bc.Head)
    return s
}

func (bc *BlockCache) setHead(block *JitBlock) {
	block.Prev = bc.Head
	block.Next = bc.Head.Next
	bc.Head.Next.Prev = block
	bc.Head.Next = block
}

func (bc *BlockCache) setTail(block *JitBlock) {
	block.Prev = bc.Tail.Prev
	block.Next = bc.Tail
	bc.Tail.Prev.Next = block
	bc.Tail.Prev = block
}

func (bc *BlockCache) remove(block *JitBlock) {
    prev := block.Prev
    next := block.Next
    prev.Next = next
    next.Prev = prev
}


func (bc *BlockCache) PopTail() *JitBlock {
    block := bc.Tail.Prev
    bc.remove(block)
    bc.setHead(block)
    return block
}

func (bc *BlockCache) PushHead(block *JitBlock) {
    bc.remove(block)
    bc.setHead(block)
}

func (bc *BlockCache) PushTail(block *JitBlock) {
    bc.remove(block)
    bc.setTail(block)
}


func (bc *BlockCache) AssignBlock(jit *Jit) *JitBlock {

    bc.mu.Lock()

    block := bc.PopTail()

    refs := block.refs.Load()
    if refs != 0 {
        println("assign refs", refs)
        return nil
    }

    block.refs.Add(1)

    if inUse := block.Length != 0; inUse {
        // need to invalidate pc currently using block
        revokedPC := block.initPc
        page := jit.Pages[revokedPC>>PAGE_SHIFT].Load()
        if page != nil {
            page.Blocks[(revokedPC&PAGE_MASK)>>2].Swap(nil)
        }
    }

    block.assembler.Off = 0
    block.initPc  = 0
    block.Length  = 0
    block.finalOp = 0
    block.Skip    = false
    block.f = nil

    //fmt.Printf("Assig %p %04d %v", block, bc.Cnt, bc.String())
    block.refs.Add(-1)

    if refs := block.refs.Load(); refs != 0 {
        panic(fmt.Sprintf("jit ref != 0 is %d", refs))
    }

    bc.mu.Unlock()
    return block
}

func (bc *BlockCache) TouchBlock(block *JitBlock) {
    bc.mu.Lock()

    refs := block.refs.Load()
    if refs != 0 {
        println("refs", refs)
        return
    }

    block.refs.Add(1)

    bc.PushHead(block)
    //fmt.Printf("Touch %p %04d %v", block, bc.Cnt, bc.String())

    block.refs.Add(-1)

    if refs := block.refs.Load(); refs != 0 {
        panic(fmt.Sprintf("jit ref != 0 is %d", refs))
    }

    bc.mu.Unlock()
}

func (bc *BlockCache) InvalidateBlock(block *JitBlock) {

    bc.mu.Lock()

    if refs := block.refs.Load(); refs != 0 {
        err := fmt.Errorf("invalidating block with refs, should not occur, refs: %d, block %p", refs, block)
        panic(err)
    }

    bc.PushTail(block)
    //fmt.Printf("Inval %p %04d %v", block, bc.Cnt, bc.String())

    bc.mu.Unlock()
}
