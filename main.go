package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// TODO(bvdberg): create repo on github
// TODO(bvdberg): add all files to repo
// TODO(bvdberg): govendor sync
// TODO(bvdberg): create wercker app (plus pipelines, env vars, etc) [staging, production]

var ErrorExitCode = cli.NewExitError("", 1)

func init() {
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	rand.Seed(time.Now().UnixNano())
}

var (
	initCommand = cli.Command{
		Name:  "init",
		Usage: "start a new project based on a template",
		Action: func(c *cli.Context) error {
			err := initAction(c)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}
			return err
		},
	}
	applyCommand = cli.Command{
		Name:  "apply",
		Usage: "update a project based on a template",
		Action: func(c *cli.Context) error {
			err := applyAction(c)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}
			return err
		},
	}
)

// initAction generates a config and then does an apply on an empty dir
func initAction(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return fmt.Errorf("Need a template and a name")
	}
	template := c.Args()[0]
	name := c.Args()[1]
	port := randomInt(1024, 65535)
	gatewayPort := port + 1
	healthPort := port + 2
	metricsPort := port + 3
	description := "I am too lazy to write a description for my project and am a bad person"

	config := &Config{
		Template:    template,
		Name:        name,
		Port:        port,
		GatewayPort: gatewayPort,
		HealthPort:  healthPort,
		MetricsPort: metricsPort,
		Description: description,
		Year:        time.Now().Format("2006"),
	}

	templatesPath := c.GlobalString("templates-path")
	managedPath := c.GlobalString("managed-path")

	// verify template exists
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

	// verify output path isn't there (since we are initializing)
	outputPath := path.Join(managedPath, name)
	outputStats, err := os.Stat(outputPath)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithFields(log.Fields{
			"outputPath": outputPath,
		}).Error("Unable to access output directory")
		return ErrorExitCode
	}

	if outputStats != nil {
		log.WithFields(log.Fields{
			"outputPath": outputPath,
		}).Error("Output directory aready exists")
		return ErrorExitCode
	}

	err = os.MkdirAll(outputPath, 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"outputPath": outputPath,
		}).Error("Unable to create output directory")
		return ErrorExitCode
	}

	return ApplyBlueprint(config, templatePath, outputPath)
}

func applyAction(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return fmt.Errorf("Need a template and a name")
	}
	template := c.Args()[0]
	name := c.Args()[1]

	templatesPath := c.GlobalString("templates-path")
	managedPath := c.GlobalString("managed-path")

	outputPath := path.Join(managedPath, name)
	config, err := loadConfig(outputPath)
	if err != nil {
		return err
	}

	// verify template exists
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

	return ApplyBlueprint(config, templatePath, outputPath)
}

func main() {
	app := cli.NewApp()

	app.Name = "blueprint"
	app.Usage = "Create and manage services"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "templates-path",
			Value: "./templates",
		},
		cli.StringFlag{
			Name:  "managed-path",
			Value: "./managed",
		},
	}
	app.Commands = []cli.Command{
		initCommand,
		applyCommand,
	}

	app.Run(os.Args)
}

func loadConfig(outputPath string) (*Config, error) {
	var cfg Config

	configPath := fmt.Sprintf("%s/.managed.json", outputPath)
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ApplyBlueprint(cfg *Config, templatePath, outputPath string) error {
	walker := createWalker(cfg, templatePath, outputPath)
	err := filepath.Walk(templatePath, walker)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"template": templatePath,
		}).Error("Unable to traverse template directory")
		return ErrorExitCode
	}
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
	gatewayPort := strconv.Itoa(portInt + 1)
	healthPort := strconv.Itoa(portInt + 2)
	metricsPort := strconv.Itoa(portInt + 3)
	ask := !c.GlobalBool("y")

	result := map[string]string{
		"Name":        name,
		"Port":        port,
		"Description": description,
		"GatewayPort": gatewayPort,
		"HealthPort":  healthPort,
		"MetricsPort": metricsPort,
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
