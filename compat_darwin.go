//go:build darwin && cgo

package main

/*
// Intentionally blank: this file exists to enable cgo for the main package on darwin builds,
// so the accompanying compat_darwin.c is compiled and linked into the final binary.
*/
import "C"

