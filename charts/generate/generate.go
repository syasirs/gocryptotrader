package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/core"
)

var (
	templatePath      string
	outputFile        string
	outputPath        string
	defaultPath       = filepath.Join("..", "templates")
	defaultOutputfile = filepath.Join(outputPath, "chart_data_gen.go")
	defaultOutputPath = filepath.Join("..")
)

var tmpl = template.Must(template.New("").Parse(`// Code Generated by chart template generation tool; DO NOT EDIT.
package charts

var (
	templateList = map[string][]byte{
{{- range .Data }}
	{{ printf "%q: %v" .Name .Data }},
{{- end }}
}
)
`))

type templateData struct {
	Name string
	Data string
}

func main() {
	fmt.Println("GoCryptoTrader: chart template generator")
	fmt.Println(core.Copyright)
	fmt.Println()

	flag.StringVar(&templatePath, "path", defaultPath, "path to load templates from")
	flag.StringVar(&outputFile, "file", defaultOutputfile, "override output file")
	flag.StringVar(&outputPath, "output", defaultOutputPath, "override output path")
	flag.Parse()

	data, err := generateMap()
	if err != nil {
		log.Fatal(err)
	}

	err = common.CreateDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	fullout := filepath.Join(outputPath, outputFile)
	f, err := os.Create(fullout)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(f, struct {
		Data []templateData
	}{
		Data: data,
	})
	defer func() {
		err = f.Close()
		if err != nil {
			log.Print(err)
		}
	}()
	if err != nil {
		log.Print(err)
	}

	cmd := exec.Command("go", "fmt")
	cmd.Dir = outputPath
	out, err := cmd.Output()
	if err != nil {
		log.Printf("unable to go fmt. output: %s err: %s", out, err)
	}

	log.Printf("Template: %v generated", fullout)
}

func buildFileList() ([]string, error) {
	var files []string
	if templatePath == "" {
		return []string{}, errors.New("no template path found")
	}
	err := filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return files, err
	}
	return files, nil
}

func readTemplateToByte(input string) ([]byte, error) {
	return ioutil.ReadFile(input)
}

func stripPath(in string) string {
	return strings.TrimPrefix(in, templatePath+string(filepath.Separator))
}

func byteJoin(b []byte) string {
	s := make([]string, 0, len(b))
	for i := range b {
		s = append(s, strconv.Itoa(int(b[i])))
	}
	return "{" + strings.Join(s, ",") + "}"
}

func generateMap() ([]templateData, error) {
	templateFileList, err := buildFileList()
	if err != nil {
		return nil, err
	}
	if len(templateFileList) == 0 {
		return nil, err
	}
	var resp []templateData
	for x := range templateFileList {
		b, err := readTemplateToByte(templateFileList[x])
		if err != nil {
			log.Printf("Unable to read template: %v Reason: %v", templateFileList[x], err)
		}
		resp = append(resp, templateData{
			Name: stripPath(templateFileList[x]),
			Data: byteJoin(b),
		})
	}
	return resp, nil
}
