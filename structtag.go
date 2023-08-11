package viper

var (
	tagName = "in"
)

// UseTag sets the tag name to parse directives.
func UseTag(tag string) {
	tagName = tag
}
