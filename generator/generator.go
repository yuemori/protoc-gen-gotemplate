package generator

import (
	"bytes"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"log"
	"os"
	"strings"
	"text/template"
)

type Generator struct {
	*bytes.Buffer
	templatePaths  []string
	allFiles       []*descriptor.FileDescriptorProto          // All files in the tree
	allFilesByName map[string]*descriptor.FileDescriptorProto // All files by filename.
	Request        *plugin.CodeGeneratorRequest               // The input.
	Response       *plugin.CodeGeneratorResponse              // The output.
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
	g.generate(g.allFiles)
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
		tpl := template.Must(template.ParseFiles(path))
		if err := tpl.Execute(g.Buffer, files); err != nil {
			g.Error(err)
		}

		g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
			Name:    String(strings.Replace(tpl.Name(), ".tmpl", "", 1)),
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

func (g *Generator) WrapTypes() {
	g.allFiles = make([]*descriptor.FileDescriptorProto, 0, len(g.Request.ProtoFile))
	g.allFilesByName = make(map[string]*descriptor.FileDescriptorProto, len(g.allFiles))
	for _, f := range g.Request.ProtoFile {
		// We must wrap the descriptors before we wrap the enums
		g.allFiles = append(g.allFiles, f)
		g.allFilesByName[f.GetName()] = f
	}
}

func (g *Generator) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-go: error:", s)
	os.Exit(1)
}

// Fail reports a problem and exits the program.
func (g *Generator) Fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-go: error:", s)
	os.Exit(1)
}
