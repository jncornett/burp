package context

type Context interface {
	Run(name string, args []string) (bool, error)
}
