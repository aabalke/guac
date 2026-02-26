package debug

//
//import (
//	"os"
//
//	"github.com/aabalke/guac/emu/nds/mem"
//)
//
//func ExportMemory(m *mem.Mem, start, end uint32, path string) {
//
//	buf := make([]uint8, end-start)
//
//	for i := range uint32(len(buf)) {
//		buf[i] = m.Read(start + i, true)
//	}
//
//    ok := writeFile(path, buf)
//    if !ok {
//        panic("failed to export memory")
//    }
//
//    println("exported memory. Exiting")
//    os.Exit(0)
//
//}
