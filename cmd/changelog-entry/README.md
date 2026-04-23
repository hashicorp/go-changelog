# changelog-entry

`changelog-entry` is a command that will generate a changelog entry based on the information passed and the information retrieved from the Github repository.

The default changelog entry template is embedded from [`changelog-entry.tmpl`](changelog-entry.tmpl) but a path to a custom template can also can be passed as parameter.

The type parameter can be one of the following:
* bug
* note
* enhancement
* new-resource
* new-datasource
* new-ephemeral
* new-list-resource
* new-function
* new-action
* deprecation
* breaking-change
* feature

## Usage

```sh
$ changelog-entry -type improvement -subcategory monitoring -description "optimize the monitoring endpoint to avoid losing logs when under high load"
```

If parameters are missing the command will prompt to fill them, the pull request number is optional and if not provided the command will try to guess it based on the current branch name and remote if the current directory is in a git repository.

### Customizing the allowed types

To customize the types that will be displayed in the prompt, create a line
delimited file with the types are allow and pass it as the `-allowed-types-file`
flag.

As an example, to allow only `bug` and `enhancement` types, create a file with the following content:

```sh
bug
enhancement
```

Then pass it to the command:

```sh
$ changelog-entry -allowed-types-file=types.txt
```

## Output

``````markdown
```release-note:improvement
monitoring: optimize the monitoring endpoint to avoid losing logs when under high load
```
``````

Any failures will be logged to stderr. The entry will be written to a file named `{PR_NUMBER}.txt`, in the current directory unless an output directory is specified.
