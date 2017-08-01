package generator

import (
	pbgen "github.com/golang/protobuf/protoc-gen-go/generator"
)

type Generator struct {
	*pbgen.Generator
	packageName      string                     // What we're calling ourselves.
	allFiles         []*FileDescriptor          // All files in the tree
	allFilesByName   map[string]*FileDescriptor // All files by filename.
	genFiles         []*FileDescriptor          // Those files we will generate output for.
	file             *FileDescriptor            // The file we are compiling now.
	usedPackages     map[string]bool            // Names of packages used in current file.
	typeNameToObject map[string]Object          // Key is a fully-qualified name in input syntax.
	init             []string                   // Lines to emit in the init function.
	indent           string
	writeOutput      bool
}

func New() *Generator {
	g := new(Generator)
	return g
}

func (g *Generator) generate(file *pbgen.FileDescriptor) {
	g.file = g.FileOf(file.FileDescriptorProto)
	g.usedPackages = make(map[string]bool)

	if g.file.index == 0 {
		// For one file in the package, assert version compatibility.
		g.P("// This is a compile-time assertion to ensure that this generated file")
		g.P("// is compatible with the proto package it is being compiled against.")
		g.P("// A compilation error at this line likely means your copy of the")
		g.P("// proto package needs to be updated.")
		g.P("const _ = ", g.Pkg["proto"], ".ProtoPackageIsVersion", generatedCodeVersion, " // please upgrade the proto package")
		g.P()
	}
	for _, td := range g.file.imp {
		g.generateImported(td)
	}
	for _, enum := range g.file.enum {
		g.generateEnum(enum)
	}
	for _, desc := range g.file.desc {
		// Don't generate virtual messages for maps.
		if desc.GetOptions().GetMapEntry() {
			continue
		}
		g.generateMessage(desc)
	}
	for _, ext := range g.file.ext {
		g.generateExtension(ext)
	}
	g.generateInitFunction()

	// Run the plugins before the imports so we know which imports are necessary.
	g.runPlugins(file)

	g.generateFileDescriptor(file)

	// Generate header and imports last, though they appear first in the output.
	rem := g.Buffer
	g.Buffer = new(bytes.Buffer)
	g.generateHeader()
	g.generateImports()
	if !g.writeOutput {
		return
	}
	g.Write(rem.Bytes())

	// Reformat generated code.
	fset := token.NewFileSet()
	raw := g.Bytes()
	ast, err := parser.ParseFile(fset, "", g, parser.ParseComments)
	if err != nil {
		// Print out the bad code with line numbers.
		// This should never happen in practice, but it can while changing generated code,
		// so consider this a debugging aid.
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(raw))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		g.Fail("bad Go source code was generated:", err.Error(), "\n"+src.String())
	}
	g.Reset()
	err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(g, fset, ast)
	if err != nil {
		g.Fail("generated Go source code could not be reformatted:", err.Error())
	}
}
