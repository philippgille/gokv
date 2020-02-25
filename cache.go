package gokv

type Cache interface {
	Store
	SetterWithExp
}
