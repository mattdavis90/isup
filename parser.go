package main

import ()

//go:generate sh -c "pigeon jobparser/grammar.peg | goimports > jobparser/grammar.go"
//go:generate sh -c "pigeon testparser/grammar.peg | goimports > testparser/grammar.go"
