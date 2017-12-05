//-----------------------------------------------------------------------------
// Copyright (c) 2017 Oracle and/or its affiliates.  All rights reserved.
// This program is free software: you can modify it and/or redistribute it
// under the terms of:
//
// (i)  the Universal Permissive License v 1.0 or at your option, any
//      later version (http://oss.oracle.com/licenses/upl); and/or
//
// (ii) the Apache License v 2.0. (http://www.apache.org/licenses/LICENSE-2.0)
//-----------------------------------------------------------------------------

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func contains(needle string, haystack []string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}

func main() {
	file := flag.String("input", "", "path to the file containing the interface")
	target := flag.String("target", "", "name of the interface to use")
	ignore := flag.String("ignore", "", "ignore the following methods (separate with commas)")
	templateFlag := flag.String("template", "", "path to the template")
	formatCode := flag.Bool("format", true, "format output using gofmt")
	output := flag.String("output", "-", "path to the output file (use - for stdout)")
	imports := flag.String("imports", "-", "custom imports (separate with commas)")

	flag.Parse()

	log.Printf("Parsing %s for interface %s", *file, *target)

	ignoredMethods := []string{}
	if *ignore != "" {
		ignoredMethods = strings.Split(*ignore, ",")
		log.Printf("Ignoring method(s): %s", strings.Join(ignoredMethods, ", "))
	}

	customImports := []string{}
	if *imports != "" {
		customImports = strings.Split(*imports, ",")
		log.Printf("Custom imports: %s", strings.Join(customImports, ", "))
	}

	tmpl, err := loadTemplate(*templateFlag)
	if err != nil {
		panic(err)
	}

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, *file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	methods := []*ast.Field{}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name != *target {
				break
			}

			i, ok := x.Type.(*ast.InterfaceType)
			if !ok {
				break
			}

			for _, m := range i.Methods.List {
				methods = append(methods, m)
			}
		}

		return true
	})

	m := []*Method{}
	for _, met := range methods {
		if t, ok := met.Type.(*ast.FuncType); ok {
			name := met.Names[0].Name

			if contains(name, ignoredMethods) {
				continue
			}

			counter := 0
			params := []*Arg{}
			for _, par := range t.Params.List {
				name := fmt.Sprintf("param%d", counter)
				if len(par.Names) > 0 {
					name = par.Names[0].Name
				}

				params = append(params, &Arg{Name: name, Type: getType(par.Type)})
				counter++
			}

			counter = 0
			returns := []*Arg{}
			for _, ret := range t.Results.List {
				name := fmt.Sprintf("result%d", counter)
				if len(ret.Names) > 0 {
					name = ret.Names[0].Name
				}

				returns = append(returns, &Arg{Name: name, Type: getType(ret.Type)})
				counter++
			}

			doclines := []string{}
			if met.Doc != nil && len(met.Doc.List) > 0 {
				for _, line := range met.Doc.List {
					doclines = append(doclines, line.Text)
				}
			}

			m = append(m, &Method{Name: name, Params: params, Returns: returns, Doc: doclines})
		}
	}

	t, err := template.New("t").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	vm := &ViewModel{
		File:          *file,
		Target:        *target,
		Template:      *templateFlag,
		Output:        *output,
		Methods:       m,
		CustomImports: customImports,
	}

	var sink bytes.Buffer
	err = t.Execute(&sink, vm)
	if err != nil {
		panic(err)
	}

	b := sink.Bytes()
	if *formatCode {
		b, err = format.Source(b)
		if err != nil {
			log.Printf("unable to format generated code")
			panic(err)
		}
	}

	var out io.Writer = os.Stdout
	if *output != "-" {
		f, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("unable to open output")
			panic(err)
		}
		defer f.Close()
		out = f
	}

	_, err = out.Write(b)
	if err != nil {
		log.Printf("unable to write output")
		panic(err)
	}
}

type ViewModel struct {
	File          string
	Target        string
	Template      string
	Output        string
	Methods       []*Method
	CustomImports []string
}

type Method struct {
	Name    string
	Doc     []string
	Params  []*Arg
	Returns []*Arg
}

type Arg struct {
	Name string
	Type string
}

func getType(n ast.Expr) string {
	switch x := n.(type) {
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", x.X, x.Sel)
	case *ast.Ident:
		return fmt.Sprintf("%s", x.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", getType(x.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", getType(x.Elt))
	}

	return fmt.Sprintf("unable to process type: %T", n)
}

func loadTemplate(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
