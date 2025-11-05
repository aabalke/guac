package debug

import (
	"bufio"
	"fmt"
	"os"
    "os/exec"
)

var L Logger

const(
    BUF_SIZE = 0xFFF
)

type Logger struct {
	file *os.File
    bufWriter *bufio.Writer
    path string
    started bool
    cnt uint64
}

func Init(path string) *Logger {

    if L.started {
        panic("STARTED LOGGER WHICH WAS ALREADY STARTED")
    }

    f, err := os.Create(path)
    if err != nil {
        panic(err)
    }

    L.path = path
    L.file = f
    L.bufWriter = bufio.NewWriter(f)

    return &L
}

func (l *Logger) Close() {

    l.bufWriter.Flush()
    l.file.Close()

    fmt.Printf("Closing Logger\n")

    var cmd *exec.Cmd
    cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", l.path)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Start()
    if err != nil {
        panic(err)
    }

    l = &Logger{}
}

func (l *Logger) Write(s string) {

    fmt.Fprintf(l.bufWriter, "%s\n", s)

    if l.cnt & BUF_SIZE == 0 {
        l.bufWriter.Flush()
    }

    l.cnt++
}
