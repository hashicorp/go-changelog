# changelog-gen

`changelog-gen` is a command that will generate a changelog entry based on the information passed and the information retrieved from the Github repository.

The changelog template by default is as follow but also can be passed as parameter:

```gotemplate
```release-note:{{ .Type}}
{{if eq .URL ""}}{{if eq .Service ""}}{{.Description}}.{{else}}{{.Service}}: {{.Description}}.{{end}}{{else}}{{if eq .Service ""}}{{.Description}}. [GH-{{.Pr}}]({{.URL}}){{else}}{{.Service}}: {{.Description}}. [GH-{{.Pr}}]({{.URL}}){{end}}{{end}}
```

The type parameter can be one of the following:
* bug
* note
* enhancement
* new-resource
* new-datasource
* deprecation
* breaking-change
* feature


## Usage

```sh
$ changelog-gen -type improvement -service monitoring -description "optimize the monitoring endpoint to avoid losing logs when under high load"
```
if parameters are missing the command will prompt to fill them, the pr number is optional and if not provided the command will try to guss it based on the current branch name (current folder)


## Results

Any failures will be logged to stderr. The entry will be written in the current folder
