package main

import (
	"bytes"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jjeffery/sqlrow/private/gen"
	"github.com/spf13/pflag"
)

var command struct {
	filename string
	output   string
}

func main() {
	log.SetFlags(0)
	command.filename = os.Getenv("GOFILE")
	pflag.StringVarP(&command.filename, "file", "f", command.filename, "source file")
	pflag.StringVarP(&command.output, "output", "o", defaultOutput(command.filename), "output")
	pflag.Parse()
	if len(pflag.Args()) > 0 {
		log.Fatalln("unrecognized args:", strings.Join(pflag.Args(), " "))
	}
	if command.filename == "" {
		log.Fatal("no file specified (-f or $GOFILE)")
	}

	model, err := gen.Parse(command.filename)
	if err != nil {
		log.Fatalln(err)
	}
	model.CommandLine = strings.Join(os.Args, " ")

	var buf bytes.Buffer
	if err := gen.DefaultTemplate.Execute(&buf, model); err != nil {
		log.Fatalln("cannot execute template:", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalln("cannot format generated output:", err)
	}

	var output io.Writer

	if command.output == "" || command.output == "-" {
		output = os.Stdout
	} else {
		outfile, err := os.Create(command.output)
		if err != nil {
			log.Fatalln(err)
		}
		defer outfile.Close()
		output = outfile
	}

	if _, err := output.Write(formatted); err != nil {
		log.Fatalln(err)
	}
}

func defaultOutput(filename string) string {
	if filename == "" {
		return ""
	}
	output := strings.TrimSuffix(filename, filepath.Ext(filename))
	output = output + "_sqlrow.go"
	return output
}