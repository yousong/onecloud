package shell

import (
	"fmt"
)

type SubParserArgs struct {
	Opts        interface{}
	Command     string
	Description string
	Callback    interface{}
}

var SubParserArgsMap = map[string]*SubParserArgs{}

func R(args *SubParserArgs) {
	argsOld, ok := SubParserArgsMap[args.Command]
	if ok {
		m := fmt.Sprintf("command %q: subparser args already registered: %#v", args.Command, argsOld)
		panic(m)
	}
	SubParserArgsMap[args.Command] = args
}
