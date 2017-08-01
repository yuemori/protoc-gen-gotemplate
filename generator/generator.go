package generator

import (
	"bytes"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	gen "github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"log"
	"strings"
	"text/template"
)

type Generator struct {
	*bytes.Buffer
	gen.Generator
	templatePaths  []string
	genFiles       []*gen.FileDescriptor          // Those files we will generate output for.
	allFiles       []*gen.FileDescriptor          // All files in the tree
	allFilesByName map[string]*gen.FileDescriptor // All files by filename.
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
	services := make([]*Service, 0, len(g.allFiles))

	for _, file := range g.allFiles {
		fdp := file.FileDescriptorProto
		for _, sv := range fdp.Service {
			service := &Service{
				Name:        sv.GetName(),
				PackageName: file.GetPackage(),
				Methods:     sv.Method}
			services = append(services, service)
		}
	}
	g.Reset()
	g.generate(services)
}

func (g *Generator) generate(services []*Service) {
	g.Buffer = new(bytes.Buffer)
	rem := g.Buffer
	g.execTemplate(services)
	g.Write(rem.Bytes())
	g.Reset()
}

func String(v string) *string {
	return &v
}

func (g *Generator) execTemplate(services []*Service) {
	for _, path := range g.templatePaths {
		tpl := template.Must(template.ParseFiles(path))
		if err := tpl.Execute(g.Buffer, services); err != nil {
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
		log.Print(p)
		g.templatePaths = append(g.templatePaths, p)
	}
}

func (g *Generator) WrapTypes() {
	g.allFiles = make([]*gen.FileDescriptor, 0, len(g.Request.ProtoFile))
	g.allFilesByName = make(map[string]*gen.FileDescriptor, len(g.allFiles))
	for _, f := range g.Request.ProtoFile {
		// We must wrap the descriptors before we wrap the enums
		fd := &gen.FileDescriptor{
			FileDescriptorProto: f,
		}
		g.allFiles = append(g.allFiles, fd)
		g.allFilesByName[f.GetName()] = fd
	}

	g.genFiles = make([]*gen.FileDescriptor, 0, len(g.Request.FileToGenerate))
	for _, fileName := range g.Request.FileToGenerate {
		fd := g.allFilesByName[fileName]
		if fd == nil {
			g.Fail("could not find file named", fileName)
		}
		g.genFiles = append(g.genFiles, fd)
	}
}
