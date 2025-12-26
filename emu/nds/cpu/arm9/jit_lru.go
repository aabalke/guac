package arm9
//
//import "fmt"
//
//func (j *Jit) add(page *Page) {
//
//	page.Prev = j.Head
//	page.Next = j.Head.Next
//
//	j.Head.Next.Prev = page
//	j.Head.Next = page
//
//	j.Pages[page.id] = page
//}
//
//func (j *Jit) remove(page *Page) {
//
//	prev := page.Prev
//	next := page.Next
//
//	prev.Next = next
//	next.Prev = prev
//
//	//fmt.Printf("PAGE %08X\n", page.id)
//
//}
//
//func (j *Jit) moveToHead(page *Page) {
//
//	if page.id == j.Head.Next.id {
//		return
//	}
//
//	j.remove(page)
//	j.add(page)
//}
//
//func (j *Jit) popTail() *Page {
//	res := j.Tail.Prev
//	j.remove(res)
//	return res
//}
//
//func (j *Jit) get(id uint32) *Page {
//
//	page := j.Pages[id]
//
//	if page == nil {
//		return nil
//	}
//
//	j.moveToHead(page)
//	return page
//}
////
////func (j *Jit) set(pc uint32) {
////
////	//fmt.Printf("PAGE IDX %08X\n", j.Head.Next.id)
////
////	pageIdx := pc >> PAGE_SHIFT
////
////	if page := j.Pages[pageIdx]; page == nil {
////
////		newPage := &Page{
////			id:     pageIdx,
////			Blocks: make([]*JitBlock, (1<<PAGE_SHIFT)>>2),
////		}
////
////		//j.Metrics[pageIdx] = make([]uint32, (1<<PAGE_SHIFT)>>2)
////
////		//fmt.Printf("PAGE %08X CNT %02d\n", pageIdx, j.Cnt)
////        fmt.Printf("ADDING %08X\n", newPage.id)
////		j.add(newPage)
////        j.GetPagesData()
////
////		j.Cnt++
////
////		if j.Cnt > j.Capacity {
////			tail := j.popTail()
////			fmt.Printf("TAIL %08X\n", tail.id)
////            j.DeletePage(tail.id)
////			j.remove(tail)
////			j.Cnt--
////            panic("COMPLETED FIRST DELETE")
////		}
////
////	} else {
////		j.moveToHead(page)
////	}
////
////	//blockIdx := (pc & PAGE_MASK) >> 2 // aligned to arm
////	//j.Metrics[pageIdx][blockIdx]++
////}
////
////
////// temp
//
//func (j *Jit) GetPagesData() {
//
//    //fmt.Printf("--- PAGE COUNT %04d\n", j.Cnt)
//
//    pages := []uint32{}
//
//    for i := range uint32(len(j.Pages)) {
//
//        if j.Pages[i] != nil {
//            pages = append(pages, i)
//            //j.GetPageData(i)
//        }
//    }
//
//    fmt.Printf("PAGES %08X\n", pages)
//}
//
//func (j *Jit) GetPageData(idx uint32) {
//
//    page := j.Pages[idx]
//
//    //blocks := []*JitBlock{}
//    blocks := []uint32{}
//
//    for i := range len(page.Blocks) {
//        if page.Blocks[i] != nil {
//            //blocks = append(blocks, page.Blocks[i])
//            blocks = append(blocks, page.Blocks[i].initPc)
//        }
//    }
//
//    fmt.Printf("PAGE %08X BLOCKS %08X\n", idx, blocks)
//
//}
