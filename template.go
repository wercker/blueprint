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
	"io/ioutil"
	"strings"
	"text/template"
)

var replacements [][]string = [][]string{
	[]string{"blueprint/templates/service", "{{lower .Name}}"},
	[]string{"Blueprint", "{{title .Name}}"},
	[]string{"blueprint", "{{lower .Name}}"},
	[]string{"blue_print", "{{packaging .Name}}"},
	[]string{"6666", "{{.Port}}"},
	[]string{"6667", "{{.GatewayPort}}"},
	[]string{"6668", "{{.HealthPort}}"},
	[]string{"6669", "{{.MetricsPort}}"},
	[]string{"TiVo for VRML", "{{.Description}}"},
	[]string{"1996", "{{.Year}}"},
}

func replaceSentinels(s string) string {
	for _, x := range replacements {
		search, replace := x[0], x[1]
		s = strings.Replace(s, search, replace, -1)
	}
	return s
}

func getTemplate(templatePath string) (*template.Template, error) {
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}
	contentString := string(content)
	contentString = replaceSentinels(contentString)

	tmpl, err := template.New(templatePath).Funcs(Funcs).Parse(contentString)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func studly(s string) string {
	parts := strings.Split(s, "-")
	newParts := []string{}
	for _, part := range parts {
		newParts = append(newParts, strings.Title(part))
	}
	return strings.Join(newParts, "")
}

func packaging(s string) string {
	return strings.Replace(s, "-", "", -1)
}

var Funcs template.FuncMap = template.FuncMap{
	// "package": func(input string) string { return strings.ToLower(input) },
	// "method":  func(input string) string { return strings.Title(input) },
	// "class":   func(input string) string { return strings.Title(input) },
	// "file":    func(input string) string { return strings.ToLower(input) },
	"title":     studly,
	"packaging": packaging,
	"lower":     func(input string) string { return strings.ToLower(input) },
}
