package tests

import "os"

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = os.Getenv("TODO_TEST_TOKEN")
