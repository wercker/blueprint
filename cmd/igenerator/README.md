# igenerator

> igenerator is short for interface generator

igenerator parses a go file and searches for specified interface. It then
passes the method information to a template and stores the result.

# Flags

```
  -format
    	format output using gofmt (default true)
  -ignore string
    	ignore the following methods (separate with comma's)
  -input string
    	path to the file containing the interface
  -output string
    	path to the output file (use - for stdout) (default "-")
  -target string
    	name of the interface to use
  -template string
    	path to the template
```

# Usage

Typical usage will be to reference the Docker image from within a repository,
and start it using `go generate`. New projects will already have this for the
metrics and trace store. Older projects can include this by copying
[templates/service/state/generate.go](../../templates/service/state/generate.go)
and
[templates/service/state/generate-stores.sh.go](../../templates/service/state/generate-stores.sh.go)
into their project.

# Docker image 

igenerator provides a Docker image which can be used to generate the output in
a stable way without bundling the igenerator code or templates within a
project. This image will have `go run cmd/igenerator/main.go` set as the
entrypoint.

You can use the `igenerator-push-quay` pipeline to generate and push a new
Docker image to a registry. Set the `$TPL_IGENERATOR_VERSION` environment
variable to change the tag. During testing make sure that this is not set to
`stable`. Only when testing is complete and a new version is deemed stable
should it use `stable`.
