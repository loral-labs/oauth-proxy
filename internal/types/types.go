package types

// define an enum for the context keys
type ContextKey string

const (
	OryClientKey   ContextKey = "OryClient"
	ConfigKey      ContextKey = "Config"
	StoreKey       ContextKey = "Store"
	BearerTokenKey ContextKey = "BearerToken"
	OryUserIDKey   ContextKey = "OryUserID"
)

// Patch Operation enum
type Operation string

const (
	Replace Operation = "replace"
	Add     Operation = "add"
	Remove  Operation = "remove"
	Move    Operation = "move"
	Copy    Operation = "copy"
	Test    Operation = "test"
)
