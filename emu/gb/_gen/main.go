package main

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"text/template"
)

var (
    flagDMG  = flag.Bool("arm9", true, "run arm9 generator")
    flagGBC  = flag.Bool("arm7", true, "run arm7 generator")
    flagPath = flag.String("path", "../", "export generated code to `path`")
)

type CpuConfig struct {
    GBC bool
}

func main() {

	flag.Parse()

    if flag.NArg() > 0 {
        flag.Usage()
        os.Exit(1)
        return
    }

    if *flagDMG {
        generate(*flagPath, CpuConfig{})
    }

    if *flagGBC {
        generate(*flagPath, CpuConfig{GBC: true})
    }
}

func generate(exportPath string, cfg CpuConfig) {

    buildImportPath := func(file string) string {
        return "./templates/" + file + ".gotmpl"
    }

    buildExportPath := func(file string) string {
        return exportPath + file + ".go"
    }

    s := "graphics_dmg"
    if cfg.GBC {
        s = "graphics_gbc"
    }

    for _, v := range [...]string{
        "graphics",
    } {
        generateFile(
            buildImportPath(v),
            buildExportPath(s),
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
