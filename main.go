package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	cli "gopkg.in/urfave/cli.v1"

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
			Value: "",
		},
		cli.StringFlag{
			Name:  "name",
			Value: "",
		},
		cli.StringFlag{
			Name:  "description",
			Value: "",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "",
		},
		cli.BoolFlag{
			Name: "y",
			// Value: "",
		},
	}
	app.Action = action

	app.Run(os.Args)
}

var action = func(c *cli.Context) error {
	output := c.GlobalString("output")
	if output == "" {
		name := c.GlobalString("name")
		if name != "" {
			output = name
		} else {
			output = "output"
		}
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

	vars := getVars(c, output)
	log.Println("Variables:")
	dumpVars(vars)

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

type question struct {
	Name       string
	Validators []Validator
}

func getVars(c *cli.Context, outputPath string) map[string]string {
	name := c.GlobalString("name")
	if name == "" {
		name = path.Base(outputPath)
	}

	description := c.GlobalString("description")
	if description == "" {
		description = "I am too lazy to write a description for my project and am a bad person"
	}

	port := c.GlobalString("port")
	if port == "" {
		port = strconv.Itoa(randomInt(1024, 65535))
	}

	portInt, _ := strconv.Atoi(port)
	gateway := strconv.Itoa(portInt + 1)
	ask := !c.GlobalBool("y")

	result := map[string]string{
		"Name":        name,
		"Port":        port,
		"Description": description,
		"Gateway":     gateway,
		"Year":        time.Now().Format("2006"),
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
			var val string
			if ask {
				if hasDefault {
					os.Stdout.WriteString(fmt.Sprintf("%s [%s]: ", q.Name, defaultValue))
				} else {
					os.Stdout.WriteString(fmt.Sprintf("%s: ", q.Name))
				}

				scanner.Scan()
				val = scanner.Text()

				//var val string
				//fmt.Fscanln(os.Stdin, &val)
			}

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
