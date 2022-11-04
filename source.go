package in

type SourceKind int

const (
	SourceStream SourceKind = iota // []byte("...")
	SourceEnv                      // os.Env
	SourceCLI
	SourceHTTP
)

type Source interface {
	Lookup(key string) (values []string, exists bool)
	Get(key string) (values []string)
}
