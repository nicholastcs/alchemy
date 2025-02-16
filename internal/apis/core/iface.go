package core

type ManifestPattern interface {
	Validate() error
}
