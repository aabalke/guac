package main

import (
	"bytes"
	"flag"
	"go/format"
	"os"
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

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, cfg); err != nil {
		panic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(exportPath, formatted, 0644); err != nil {
		panic(err)
	}
}
