package generator

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	*bytes.Buffer
	templatePaths []string
	Request       *plugin.CodeGeneratorRequest  // The input.
	Response      *plugin.CodeGeneratorResponse // The output.
}

type Service struct {
	Name        string
	PackageName string
	Methods     []*descriptor.MethodDescriptorProto
}

func New() *Generator {
	g := new(Generator)
	g.Buffer = new(bytes.Buffer)
	g.Request = new(plugin.CodeGeneratorRequest)
	g.Response = new(plugin.CodeGeneratorResponse)
	return g
}

func (g *Generator) GenerateAllFiles() {
	g.Reset()

	allFiles := make([]*descriptor.FileDescriptorProto, 0, len(g.Request.ProtoFile))
	for _, f := range g.Request.ProtoFile {
		allFiles = append(allFiles, f)
	}

	g.generate(allFiles)
}

func (g *Generator) generate(files []*descriptor.FileDescriptorProto) {
	g.Buffer = new(bytes.Buffer)
	rem := g.Buffer
	g.execTemplate(files)
	g.Write(rem.Bytes())
	g.Reset()
}

func String(v string) *string {
	return &v
}

func (g *Generator) execTemplate(files []*descriptor.FileDescriptorProto) {
	for _, path := range g.templatePaths {
		tplName := strings.Replace(filepath.Base(path), ".tmpl", "", 1)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			g.Error(err)
		}
		s := string(b)
		tpl := template.Must(template.New(tplName).Funcs(sprig.TxtFuncMap()).Parse(s))
		if err := tpl.Execute(g.Buffer, files); err != nil {
			g.Error(err)
		}

		g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
			Name:    String(tplName),
			Content: String(g.String()),
		})
	}
}

func (g *Generator) CommandLineParameters(parameter string) {
	paths := strings.Split(parameter, ",")
	g.templatePaths = make([]string, 0, len(paths))
	for _, p := range paths {
		g.templatePaths = append(g.templatePaths, p)
	}
}

func (g *Generator) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-go: error:", s)
	os.Exit(1)
}

func (g *Generator) Fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-go: error:", s)
	os.Exit(1)
}
