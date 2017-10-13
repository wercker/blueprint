# blueprint

Blueprint takes a template and turns it into a application. And more!

# Design

This is how a new project will be set up, we'll need to make existing projects
look enough like this for the rest of the magic to work.

file structure from blueprint root:

```
- templates           # where templates are held
  - service           # a template named service
- managed             # where services managed by blueprint live
  - inspector         # a service managed by blueprint
    - .managed.json   # the config used to generate this service
```

# Usage

Start a new project:

`blueprint init service $name`

# Generating files

We use go generate

```
govendor generate +local
```

See [cmd/igenerator/README.md](cmd/igenerator/README.md) for more information
regarding generating stores and deploying the related Docker image.

# How to Write Templates

We're using some special sentinel values that we will find replace in the doc
before doing the templating:

```
"blueprint/templates/service" => "{{lower .Name}}"},
"Blueprint" => "{{title .Name}}"},
"blueprint" => "{{lower .Name}}"},
"blue_print" => "{{packaging .Name}}"},
"6666" => "{{.Port}}"},
"6667" => "{{.GatewayPort}}"},
"6668" => "{{.HealthPort}}"},
"6669" => "{{.MetricsPort}}"},
"TiVo for VRML" => "{{.Description}}"},
"1996" => "{{.Year}}"},
}
```

Of note is blue_print, which must be used for things that are package names
