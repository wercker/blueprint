# blueprint

Blueprint takes a template and turns it into source code. And more!

# Usage

To start a project use the `init` command. It requires a couple of arguments,
such as the template and the new name of the service.

```
blueprint init [service name] [name]
```

Blueprint will prompt for these, but it is possible to provide these through
flags.

# Building

We use dep to ensure all dependencies with their expected versions are
present in the vendor directory.

```
dep ensure
go generate ./...
dep ensure
go install
```

# iGenerator

We use igenerator to generate trace and metrics stores for certain templates.
See [cmd/igenerator/README.md](cmd/igenerator/README.md) for more information
regarding generating stores and deploying the related Docker image.

# Templates

## Creating new templates

To create a new template in the templates folder. The files in there
will be expanded using using normal go templates. However we use a couple of
sentinels to make it easier to work with the templates:

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

Of note is `blue_print`, which must be used for things that are package names.

Blueprint will also replace names in the directory names and filenames.

# Future features

The following is a list of features that we want to create, but they currently
do not exist yet.

- Allow remote templates from Github or tarball's.
- Allow templates to specify their own parameters, sentinel, etc.
- Allow files to be ignored and/or sentinels in specific places to be ignored.
- Go format all .go files.

# License

Copyright (c) 2017-2018 Oracle and/or its affiliates.  All rights reserved.

This program is free software: you can modify it and/or redistribute it under
the terms of:

(i)  the Universal Permissive License v 1.0 or at your option, any
     later version (<http://oss.oracle.com/licenses/upl>); and/or

(ii) the Apache License v 2.0. (<http://www.apache.org/licenses/LICENSE-2.0>)
