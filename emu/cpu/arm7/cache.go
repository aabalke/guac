package arm7

// block cache is lru cache that assigns and returns blocks to be used by jit

import (
	"fmt"
	"strings"

	"github.com/aabalke/gojit"
)

type BlockCache struct {
	Head, Tail *JitBlock
	Blocks     []*JitBlock
	SkipBlock  *JitBlock
}

type JitBlock struct {
	f          func()
	assembler  *gojit.Assembler
	Prev, Next *JitBlock
	Thumb      bool
	Skip       bool
	initPc     uint32
	Size       uint32
}

func InitBlockCache(capacity uint32, pageSize int) *BlockCache {
	bc := &BlockCache{
		Blocks:    []*JitBlock{},
		Head:      &JitBlock{},
		Tail:      &JitBlock{},
		SkipBlock: &JitBlock{Skip: true},
	}

	for range capacity {
		bc.Blocks = append(bc.Blocks, &JitBlock{})
	}

	// head and tail are never directly used, just pivots for
	// head prev and tail next
	bc.Head.Next = bc.Tail
	bc.Tail.Prev = bc.Head

	// TODO: instead of init all same page size, can we mmap once and have variable assembler sizes?
	// TODO: confirm page_size is reasonable and not insanely overkill

	for i := range len(bc.Blocks) {

		bc.Blocks[i].Prev = bc.Head
		bc.Blocks[i].Next = bc.Head.Next

		bc.Head.Next.Prev = bc.Blocks[i]
		bc.Head.Next = bc.Blocks[i]

		asm, err := gojit.New(pageSize)
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

func (bc *BlockCache) Close() {
	for i := range len(bc.Blocks) {
		bc.Blocks[i].assembler.Release()
	}
}

func (bc *BlockCache) String() string {
	// string, tail -> head pointers
	var sb strings.Builder

	fmt.Fprintf(&sb, "B ")
	for b := bc.Tail; ; {
		if b != bc.Tail && b != bc.Head {
			fmt.Fprintf(&sb, "%p ", b)
		}

		b = b.Prev
		if b.Prev == nil {
			break
		}
	}

	fmt.Fprintf(&sb, "\n")
	return sb.String()
}

func (bc *BlockCache) BlockString() string {
	// string, need tail and head included to match String()

	var sb strings.Builder

	sb.Grow(5 + 20*len(bc.Blocks) + 20)

	fmt.Fprintf(&sb, "B %p ", bc.Tail)
	for i := range len(bc.Blocks) {
		fmt.Fprintf(&sb, "%p ", bc.Blocks[i])
	}

	fmt.Fprintf(&sb, "%p\n", bc.Head)
	return sb.String()
}

//go:inline
func (bc *BlockCache) setHead(block *JitBlock) {
	block.Prev = bc.Head
	block.Next = bc.Head.Next
	bc.Head.Next.Prev = block
	bc.Head.Next = block
}

//go:inline
func (bc *BlockCache) setTail(block *JitBlock) {
	block.Prev = bc.Tail.Prev
	block.Next = bc.Tail
	bc.Tail.Prev.Next = block
	bc.Tail.Prev = block
}

//go:inline
func (bc *BlockCache) remove(block *JitBlock) {
	prev := block.Prev
	next := block.Next
	prev.Next = next
	next.Prev = prev
}

//go:inline
func (bc *BlockCache) PopTail() *JitBlock {
	block := bc.Tail.Prev
	bc.remove(block)
	bc.setHead(block)
	return block
}

//go:inline
func (bc *BlockCache) PushHead(block *JitBlock) {
	bc.remove(block)
	bc.setHead(block)
}

//go:inline
func (bc *BlockCache) PushTail(block *JitBlock) {
	bc.remove(block)
	bc.setTail(block)
}

func (bc *BlockCache) AssignBlock(jit *Jit) *JitBlock {
	block := bc.PopTail()

	if inUse := block.Size != 0; inUse {
		// need to invalidate pc currently using block
		revokedPC := block.initPc
		page := jit.Pages[revokedPC>>jit.Config.PageShift]
		if page != nil {
			page.Blocks[(revokedPC&jit.Config.PageMask)>>1] = nil
		}
	}

	block.assembler.Off = 0
	block.initPc = 0
	block.Size = 0
	block.f = nil

	return block
}

//go:inline
func (bc *BlockCache) TouchBlock(block *JitBlock) {
	bc.PushHead(block)
}

//go:inline
func (bc *BlockCache) InvalidateBlock(block *JitBlock) {
	bc.PushTail(block)
}
