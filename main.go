package main

import (
	"bufio"
	"bytes"
	"fmt"
	"text/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	log "github.com/Sirupsen/logrus"
)

var ErrorExitCode = cli.NewExitError("", 1)

func init() {
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	rand.Seed(time.Now().UnixNano())
}

func main() {
	app := cli.NewApp()

	app.Name = "blueprint"
	app.Usage = "Create new services"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "templates-path",
			Value: "./templates",
		},
		cli.StringFlag{
			Name: "template",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "./output",
		},
	}
	app.Action = action

	app.Run(os.Args)
}

var action = func(c *cli.Context) error {
	output := c.GlobalString("output")
	if output == "" {
		log.Error("output is required")
		return ErrorExitCode
	}

	template := c.GlobalString("template")
	if template == "" {
		log.Error("template is required")
		return ErrorExitCode
	}

	templatesPath := c.GlobalString("templates-path")
	templatePath := path.Join(templatesPath, template)

	templateStats, err := os.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cli.NewExitError(fmt.Sprintf("Template not found: %s", template), 1)
		}

		return cli.NewExitError(fmt.Sprintf("Unable to access template: %+v", err), 1)
	}

	if !templateStats.IsDir() {
		return cli.NewExitError("Template is not a directory", 1)
	}

	outputStats, err := os.Stat(output)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithFields(log.Fields{
			"outputPath": output,
		}).Error("Unable to access output directory")
		return ErrorExitCode
	}

	if outputStats != nil {
		log.WithFields(log.Fields{
			"outputPath": output,
		}).Error("Output directory aready exists")
		return ErrorExitCode
	}

	err = os.MkdirAll(output, 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"outputPath": output,
		}).Error("Unable to create output directory")
		return ErrorExitCode
	}

	log.WithFields(log.Fields{
		"templatesPath": templatesPath,
		"template":      template,
	}).Debug("Traversing template directory")

	vars := getVars(output)

	log.Println("Variables:")
	dumpVars(vars)

	//fmt.Fprint(os.Stdout, "Continue? [Y/n]")
	// TODO(bvdberg): read from stdin and check if it is Y or empty

	walker := createWalker(templatePath, output, vars)
	err = filepath.Walk(templatePath, walker)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"template": template,
		}).Error("Unable to traverse template directory")
		return ErrorExitCode
	}

	// TODO(bvdberg): create repo on github
	// TODO(bvdberg): add all files to repo
	// TODO(bvdberg): govendor sync
	// TODO(bvdberg): create wercker app (plus pipelines, env vars, etc) [staging, production]

	return nil
}

func createWalker(templateRoot, outputRoot string, vars map[string]string) func(string, os.FileInfo, error) error {
	return func(p string, info os.FileInfo, err error) error {
		relativePath := strings.TrimPrefix(p, templateRoot)
		if relativePath == "" {
			return nil
		}

		templatePath := path.Join(templateRoot, relativePath)

		tmpl, err := template.
			New(path.Join(outputRoot, relativePath)).
			Funcs(Funcs).
			Parse(path.Join(outputRoot, relativePath))
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, vars)
		if err != nil {
			return err
		}
		outputPath := buf.String()

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"path": p,
			}).Error("Unable to access file")
		} else if info.IsDir() {
			return handleDirectory(templatePath, outputPath)
		} else {
			return handleFile(templatePath, outputPath, vars)
		}
		return nil
	}
}

func handleDirectory(templatePath, outputPath string) error {
	log.WithFields(log.Fields{
		"path": outputPath,
	}).Debug("Creating directory")
	return os.Mkdir(outputPath, 0777)
}

func handleFile(templatePath, outputPath string, vars map[string]string) error {
	log.WithFields(log.Fields{
		"path": outputPath,
	}).Debug("Expanding template")

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Make every .sh file executable
	if path.Ext(outputPath) == ".sh" {
		stat, err := f.Stat()
		if err != nil {
			return err
		}

		err = f.Chmod(stat.Mode() | 0700)
		if err != nil {
			return err
		}
	}

	tmpl, err := getTemplate(templatePath)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, vars)
}

func getTemplate(templatePath string) (*template.Template, error) {
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(templatePath).Funcs(Funcs).Parse(string(content))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

type question struct {
	Name       string
	Validators []Validator
}

func getVars(output string) map[string]string {
	result := map[string]string{
		"Name": path.Base(output),
		"Port": strconv.Itoa(randomInt(1024, 65535)),
		"Year": time.Now().Format("2006"),
	}

	questions := []question{
		{"Name", []Validator{Required, NoSpaces}},
		{"Description", []Validator{Required}},
		{"Port", []Validator{ValidNonPrivilegedPort}},
		{"Year", []Validator{Integer}},
	}

	scanner := bufio.NewScanner(os.Stdin)

	for _, q := range questions {
		defaultValue, hasDefault := result[q.Name]

		isInvalid := true

	ValidLoop:
		for isInvalid {
			if hasDefault {
				os.Stdout.WriteString(fmt.Sprintf("%s [%s]: ", q.Name, defaultValue))
			} else {
				os.Stdout.WriteString(fmt.Sprintf("%s: ", q.Name))
			}

			scanner.Scan()
			val := scanner.Text()

			//var val string
			//fmt.Fscanln(os.Stdin, &val)

			if val == "" {
				val = defaultValue
			}

			for _, v := range q.Validators {
				err := v.Validate(val)
				if err != nil {
					os.Stdout.WriteString(fmt.Sprintf("Invalid value: %s\n", err))
					continue ValidLoop
				}
			}

			isInvalid = false
			result[q.Name] = val
		}
	}

	return result
}

var (
	Integer                = &IntegerValidator{}
	NoSpaces               = &NoSpacesValidator{}
	Required               = &RequiredValidator{}
	ValidNonPrivilegedPort = &ValidPortValidator{true}
	ValidPort              = &ValidPortValidator{false}
)

type Validator interface {
	Validate(val string) error
}

type RequiredValidator struct{}

func (v *RequiredValidator) Validate(val string) error {
	if val == "" {
		return fmt.Errorf("Value is required")
	}
	return nil
}

type NoSpacesValidator struct{}

func (v *NoSpacesValidator) Validate(val string) error {
	if strings.Contains(val, " ") {
		return fmt.Errorf("Value cannot contain a space")
	}
	return nil
}

type ValidPortValidator struct {
	onlyNonPrivilegedPorts bool
}

func (v *ValidPortValidator) Validate(val string) error {
	minPort := int64(1)
	maxPort := int64(65535)

	if v.onlyNonPrivilegedPorts {
		minPort = 1024
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid port, requires %d-%d", minPort, maxPort)
	}

	if i < minPort || i > maxPort {
		return fmt.Errorf("Invalid port, requires %d-%d", minPort, maxPort)
	}

	return nil
}

type IntegerValidator struct {
	onlyNonPrivilegedPorts bool
}

func (v *IntegerValidator) Validate(val string) error {
	_, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func dumpVars(vars map[string]string) {
	keys := make([]string, len(vars))

	i := 0
	for name, _ := range vars {
		keys[i] = name
		i++
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := vars[key]

		fmt.Fprintf(os.Stdout, "%15s: %s\n", key, value)
	}
}

func randomInt(min, max int) int {
	i := rand.Int31n(int32(max - min))
	return int(i) + min
}

var Funcs template.FuncMap = template.FuncMap{
	"package": func(input string) string { return strings.ToLower(input) },
	"method":  func(input string) string { return strings.Title(input) },
	"class":   func(input string) string { return strings.Title(input) },
	"file":    func(input string) string { return strings.ToLower(input) },
	"title":   func(input string) string { return strings.ToLower(input) },
}
