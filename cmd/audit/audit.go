package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"unicode/utf8"
)

const RED = "\033[0;31m"
const NC = "\033[0m"
const WHITE = "\033[1;37m"
const GREEN = "\033[1;32m"

func red(s string, a ...interface{}) string {
	s = fmt.Sprintf("%s%s%s", RED, s, NC)
	return fmt.Sprintf(s, a...)
}

func white(s string, a ...interface{}) string {
	s = fmt.Sprintf("%s%s%s", WHITE, s, NC)
	return fmt.Sprintf(s, a...)
}

func green(s string, a ...interface{}) string {
	s = fmt.Sprintf("%s%s%s", GREEN, s, NC)
	return fmt.Sprintf(s, a...)
}

func failed() {
	fmt.Printf(red("✗\n"))
}

func succeeded() {
	fmt.Printf(green("✓\n"))
}

var SkipFile = fmt.Errorf("skip file")

type Wakka interface {
	Call(string, os.FileInfo, error) error
	Wrap(Wakka) Wakka
}

type WakkaFlocka struct {
	walkFunc filepath.WalkFunc
}

func (w *WakkaFlocka) Call(path string, info os.FileInfo, err error) error {
	return w.walkFunc(path, info, err)
}

func (w *WakkaFlocka) Wrap(wf Wakka) Wakka {
	var flame = func(path string, info os.FileInfo, err error) error {
		res := w.walkFunc(path, info, err)
		if res != nil {
			if res == SkipFile {
				return nil
			}
			return res
		}
		return wf.Call(path, info, err)
	}
	return &WakkaFlocka{flame}
}

func WakkaMatch(re *regexp.Regexp, only bool) Wakka {
	var flame = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		match := re.MatchString(path)
		// Skip things that don't match in only-mode
		if !match && only {
			return SkipFile
		}
		// Skip things that do match in exclude-mode
		if match && !only {
			// Skip entire matching directories
			if info.IsDir() {
				return filepath.SkipDir
			}
			return SkipFile
		}
		return nil
	}
	return &WakkaFlocka{flame}
}

func WakkaExclude(s string) Wakka {
	return WakkaMatch(regexp.MustCompile(s), false)
}

func WakkaInclude(s string) Wakka {
	return WakkaMatch(regexp.MustCompile(s), true)
}

func WakkaGrep(re *regexp.Regexp, out *[]Grepped) Wakka {
	var flame = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		g, err := Grep(path, re)
		if err != nil {
			return err
		}

		*out = append(*out, g...)
		return nil
	}
	return &WakkaFlocka{flame}
}

func Grep(path string, re *regexp.Regexp) ([]Grepped, error) {
	out := []Grepped{}
	f, err := os.Open(path)
	if err != nil {
		return out, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		if i == 0 && !utf8.ValidString(line) {
			return []Grepped{}, nil
		}
		if re.MatchString(line) {
			out = append(out, Grepped{
				Index: i,
				Line:  line,
				Match: line,
				Path:  path,
			})
		}
		i++
	}
	return out, nil
}

type Grepped struct {
	Index int
	Line  string
	Match string
	Path  string
}

func PrintGrep(grepped []Grepped) {
	last := ""
	for _, g := range grepped {
		if g.Path != last {
			fmt.Printf("%s\n", g.Path)
			last = g.Path
		}
		fmt.Printf("%d: %s\n", g.Index, g.Line)
	}
}

type ManagedJSON struct {
	Template    string
	Name        string
	Port        int
	Gateway     int
	Year        string
	Description string
}

func LoadManagedJSON(path, f string) (*ManagedJSON, error) {
	var managed ManagedJSON
	b, err := ioutil.ReadFile(filepath.Join(path, f))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &managed)
	if err != nil {
		return nil, err
	}
	return &managed, nil
}

// High Level API
func CheckNotUsing(path, s string, middleware ...Wakka) bool {
	fmt.Printf(white("Checking not using %s... ", s))
	grepped := []Grepped{}
	walker := WakkaGrep(regexp.MustCompile(s), &grepped)

	for _, wrapper := range middleware {
		walker = wrapper.Wrap(walker)
	}

	defaultFilter := WakkaExclude(`vendor|\.wercker|\.git`)
	walker = defaultFilter.Wrap(walker)

	err := filepath.Walk(path, walker.Call)
	if err != nil {
		fmt.Printf(err.Error())
	}

	if len(grepped) == 0 {
		succeeded()
		return true
	}
	failed()
	PrintGrep(grepped)
	return false
}

func CheckUsing(path, s string, middleware ...Wakka) bool {
	fmt.Printf(white("Checking using %s... ", s))
	grepped := []Grepped{}
	walker := WakkaGrep(regexp.MustCompile(s), &grepped)

	for _, wrapper := range middleware {
		walker = wrapper.Wrap(walker)
	}

	defaultFilter := WakkaExclude(`vendor|\.wercker|\.git`)
	walker = defaultFilter.Wrap(walker)

	err := filepath.Walk(path, walker.Call)
	if err != nil {
		fmt.Printf(err.Error())
	}

	if len(grepped) != 0 {
		succeeded()
		return true
	}
	failed()
	PrintGrep(grepped)
	return false
}

func CheckHasDeps(path, s string) bool {
	fmt.Printf(white("Checking for dependency %s... ", s))
	re := regexp.MustCompile(s)
	g, err := Grep(filepath.Join(path, "vendor/vendor.json"), re)
	if err != nil {
		failed()
		fmt.Printf("Error: %s\n", err.Error())
		return false
	}
	if len(g) == 0 {
		failed()
		fmt.Printf("Did not find %s in vendor/vendor.json\n", s)
		return false
	}
	succeeded()
	return true
}

func CheckNotHas(path, s string) bool {
	fmt.Printf(white("Checking not has %s... ", s))
	b, _ := filepath.Glob(filepath.Join(path, s))
	if len(b) != 0 {
		failed()
		for _, f := range b {
			fmt.Printf("Found file: %s\n", f)
		}
		return false
	}
	succeeded()
	return true
}

func CheckHas(path, s string) bool {
	fmt.Printf(white(fmt.Sprintf("Checking for %s... ", s)))
	b, _ := filepath.Glob(filepath.Join(path, s))
	if len(b) == 0 {
		failed()
		fmt.Printf("Did not find file: %s\n", s)
		return false
	}
	succeeded()
	return true
}

// Custom

// Make sure the wercker.yml is outputing the built artifacts for easy download
func CheckArtifactOutput(path string, managed *ManagedJSON) bool {
	fmt.Printf(white("Checking wercker.yml `go build` for artifact output... "))
	restring := fmt.Sprintf(`cp -r "\$WERCKER_OUTPUT_DIR/%s" "\$WERCKER_REPORT_ARTIFACTS_DIR"`, managed.Name)
	re := regexp.MustCompile(restring)

	g, err := Grep(filepath.Join(path, "wercker.yml"), re)
	if err != nil {
		failed()
		fmt.Printf("Error: %s\n", err.Error())
		return false
	}
	if len(g) == 0 {
		failed()
		fmt.Printf("Did not find artifact copy to output dir, expecting:\n%s", restring)
		return false
	}
	succeeded()
	return true
}

// // CheckTemplateVars makes sure our template variables start with WERCKER or TPL
// func CheckTemplateVars(path string) bool {

// }

func main() {
	path := os.Args[1]

	CheckNotHas(path, "glide.*")
	CheckNotUsing(path, `"github\.com/Sirupsen/logrus"`)
	CheckNotUsing(path, `"github.com/codegangsta/cli"`)
	CheckNotUsing(path, `\(c\) 2016`, WakkaExclude(".*.json"))

	// Old code
	CheckNotUsing(path, `Applying context hack to gateway`)
	CheckNotUsing(path, `func ParseObjectID(id string) (bson.ObjectId, error) {`)
	CheckHasDeps(path, `github.com/wercker/pkg/log`)

	// Only allow ${WERCKER_ and ${TPL_ in templates
	// Technically there are a few words that might satisfy this but...
	CheckNotUsing(path, `\$\{[^TW][^PE][^LR]`, WakkaInclude(`.*\.template.*`))

	// Flags flags flags flags flags env env var var var
	CheckNotUsing(path, `KUBERNETES_MASTER`)
	CheckNotUsing(path, `DOCKER_USER$`)
	CheckNotUsing(path, `MONGODB[^_]`)
	CheckNotUsing(path, `AUTH_TARGET|AuthTarget`)
	CheckNotUsing(path, `\w-host`)

	// CheckNotUsing(path, `"github\.com/wercker/pkg/log"`)

	CheckHas(path, "core/generate-protobuf.sh")
	CheckHas(path, ".managed.json")
	CheckHas(path, "version.go")
	CheckHas(path, "deployment/deployment.template.yml")

	CheckUsing(path, "kubernetes.io/change-cause")

	managed, err := LoadManagedJSON(path, ".managed.json")
	if err != nil {
		fmt.Printf(red("No .managed.json, exiting...\n"))
		os.Exit(1)
	}

	CheckArtifactOutput(path, managed)
}
