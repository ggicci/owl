# Owl

Owl is a Go package that provides algorithmic guidance through structured directives using Go struct tags.

[![Go](https://github.com/ggicci/owl/actions/workflows/go.yaml/badge.svg)](https://github.com/ggicci/owl/actions/workflows/go.yaml)
[![codecov](https://codecov.io/gh/ggicci/owl/graph/badge.svg?token=YU7FGGOY60)](https://codecov.io/gh/ggicci/owl)
[![Go Report Card](https://goreportcard.com/badge/github.com/ggicci/owl)](https://goreportcard.com/report/github.com/ggicci/owl)
[![Go Reference](https://pkg.go.dev/badge/github.com/ggicci/owl.svg)](https://pkg.go.dev/github.com/ggicci/owl)

**owl** helps you implement functionalites which require leveraging [Go struct tags](https://go.dev/wiki/Well-known-struct-tags), through:

1. owl has defined a ["standard _syntax_"](#concepts) to compose the tags, so you don't need to write code to parse tags;
2. owl can scan the struct instance (recursively) for you, so you don't need to write code to iterate the struct by yourself, which requires a lot of **reflection** operations;
3. you only need to write code to customize the behaviours of the tags, owl will call them during the scanning.

If you're still confused what **owl** can help you achieve, take a look at this tiny package: [ggicci/goenv](https://github.com/ggicci/goenv).

## Successful Customer Use Cases

| Customer                                   | Achievement                                                                                |
| ------------------------------------------ | ------------------------------------------------------------------------------------------ |
| [httpin](https://github.com/ggicci/httpin) | **owl** helps httpin enable the mutual conversion between an HTTP request and a Go struct. |

## Concepts

### Tag

Read [Go struct tags](https://go.dev/wiki/Well-known-struct-tags) if you have no idea what it is.

**owl** has a default tag named `owl`, like `json` for `encoding/json` package. The default means without any modifications, owl will only extract the `owl` tag that defined in the struct tags and parse it. You can use the following code to use any tag name you want:

```go
owl.UseTag("mytag")
```

For example, in `httpin` package, they use `in` as the tag name.

### Directive

```go
type Authorization struct {
	Token string `in:"query=access_token,token;header=x-api-token;required"`
	                  ^----------------------^ ^----------------^ ^------^
	                            d1                    d2            d3
}
```

In owl, a _directive_ is a formatted string, consisting of two parts, the [_Directive Executor_](#irective-executor) and the _arguments_, separated by an equal sign (`=`), formatted as:

```
DirectiveExecutorName=ARGV
```

Which works like a concrete function call.

To the left of the `=` is the name of the directive. There's a corresponding directive executor (with the same name) working under the hood.

To the right of the `=` are the arguments, which will be passed to the algorithm at runtime. The way to define arguments can differ across different directives. In general, comma (`,`) separated strings are used for multiple arguments. Arguments can be ommited, i.e. no `=` and the right part when defining a directive.

For the above example (`Authorization.Token`), there are three directives:

- d1: `query=access_token,token`
- d2: `header=x-api-token`
- d3: `required`

Let's dissect d1, the name of the directive is `query`, argv is `access_token,token`.

### Directive Executor

A _Directive Executor_ is an algorithm with runtime context. It is responsible for executing a concrete [directive](#directive).

**For better understanding, we can think of a _Directive Executor_ as a function in a programming language, and a _Directive_ as a concrete function call.**

| Directive | Executor                         | Directive Execution                        |
| --------- | -------------------------------- | ------------------------------------------ |
| query     | query=access_token,token         | `query(["access_token", "token"])`         |
| header    | header=x-api-token,Authorization | `header(["x-api-token", "Authorization"])` |
| required  | required                         | `required([])`                             |

## How to use (demo)?

Let's take a look at the following snippet to see how [ggicci/goenv](https://github.com/ggicci/goenv) is implemented by only a little effort with the help of owl:

```go
func exeEnvReader(rtm *owl.DirectiveRuntime) error {
	if len(rtm.Directive.Argv) == 0 {
		return nil
	}
	if value, ok := os.LookupEnv(rtm.Directive.Argv[0]); ok {
		rtm.Value.Elem().SetString(value)
	}
	return nil
}

func init() {
    owl.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))
}
```

Now, you are able to populate a struct instance by running the following code:

```go
type EnvConfig struct {
    Workspace string `owl:"env=OWL_HOME"`
    User      string `owl:"env=OWL_USER"`
    Debug     string `owl:"env=OWL_DEBUG"`
}

resolver, err := owl.New(EnvConfig{})
config, err := resolver.Resolve()
// Now, config.Workspace has the value of $OWL_HOME, ...
```
