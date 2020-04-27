package server

import "aletheia.icu/broccoli/fs"

// Mock the embedded filesystem for unit tests.
var br = fs.New(false, []byte(""))
