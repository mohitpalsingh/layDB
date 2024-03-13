package main

type DataType = string

const (
	String DataType = "String"
)

const (
	StringRecord uint64 = iota
)

// operations on string
const (
	StringSet uint64 = iota
	StringRem
)
