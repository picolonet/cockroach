package main

import (
	"github.com/cockroachdb/cockroach/pkg/picolo"
	_ "github.com/cockroachdb/cockroach/pkg/ui/distoss" // web UI init hooks
)

func main() {
	picolo.Start()
}
