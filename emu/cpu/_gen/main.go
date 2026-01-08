package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"text/template"
)

var (
    flagArm7 = flag.Bool("arm9", true, "run arm9 generator")
    flagArm9 = flag.Bool("arm7", true, "run arm7 generator")
    flagPath = flag.String("path", "../", "export generated code to `path`")
)

type CpuConfig struct {
    A9 bool
}

func main() {

	flag.Parse()

    if flag.NArg() > 0 {
        flag.Usage()
        os.Exit(1)
        return
    }

    if *flagArm7 {
        generate(*flagPath + "arm7/", CpuConfig{})
    }

    if *flagArm9 {
        generate(*flagPath + "arm9/", CpuConfig{A9: true})
    }

    if err := format(*flagPath + "arm7/..."); err != nil {
        fmt.Printf("Could not format arm7 cpu implimentation: %s\n", err.Error())
    }

    if err := format(*flagPath + "arm9/..."); err != nil {
        fmt.Printf("Could not format arm9 cpu implimentation: %s\n", err.Error())
    }

}

func generate(exportPath string, cfg CpuConfig) {

    buildImportPath := func(file string) string {
        return "./templates/" + file + ".gotmpl"
    }

    buildExportPath := func(file string) string {
        return exportPath + file + ".go"
    }

    for _, v := range [...]string{
        "arm",
        "arm_decoder",
        "arm_jit",
        "cpu",
        "exceptions",
        "jit",
        "thumb",
        "thumb_decoder",
    } {
        generateFile(
            buildImportPath(v),
            buildExportPath(v),
            cfg,
        )
    }
}

func generateFile(templatePath, exportPath string, cfg CpuConfig) {

    tmpl := template.Must(
        template.ParseFiles(templatePath),
    )

    f, err := os.Create(exportPath)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    if err := tmpl.Execute(f, cfg); err != nil {
        panic(err)
    }
}

func format(s string) error {

    goExec, err := exec.LookPath("go")
    if err != nil {
        return err
    }

    err = syscall.Exec(goExec, []string{"fmt", s}, os.Environ())
    return err
}
