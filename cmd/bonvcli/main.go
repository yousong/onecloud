package main

import (
	"context"
	"fmt"
	"os"

	"yunion.io/x/log"
	"yunion.io/x/pkg/util/version"
	"yunion.io/x/structarg"

	"yunion.io/x/onecloud/pkg/bonv/client/shell"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type Options struct {
	OsUsername string `default:"$OS_USERNAME" help:"Username, defaults to env[OS_USERNAME]"`
	OsPassword string `default:"$OS_PASSWORD" help:"Password, defaults to env[OS_PASSWORD]"`
	// OsProjectId string `default:"$OS_PROJECT_ID" help:"Proejct ID, defaults to env[OS_PROJECT_ID]"`
	OsProjectName  string `default:"$OS_PROJECT_NAME" help:"Project name, defaults to env[OS_PROJECT_NAME]"`
	OsDomainName   string `default:"$OS_DOMAIN_NAME" help:"Domain name, defaults to env[OS_DOMAIN_NAME]"`
	OsAuthURL      string `default:"$OS_AUTH_URL" help:"Defaults to env[OS_AUTH_URL]"`
	OsRegionName   string `default:"$OS_REGION_NAME" help:"Defaults to env[OS_REGION_NAME]"`
	OsZoneName     string `default:"$OS_ZONE_NAME" help:"Defaults to env[OS_ZONE_NAME]"`
	OsEndpointType string `default:"$OS_ENDPOINT_TYPE|internalURL" help:"Defaults to env[OS_ENDPOINT_TYPE] or internalURL" choices:"publicURL|internalURL|adminURL"`

	Help       bool
	Version    bool
	SUBCOMMAND string `subcommand:"true"`
}

var opts *Options = &Options{}

func prepareParser() (*structarg.ArgumentParser, error) {
	p, err := structarg.NewArgumentParser(
		opts,
		"bonvcli",
		"Command line interface to bonv API server",
		`See "bonvclient help COMMAND" for help on specific command`,
	)
	if err != nil {
		return nil, err
	}

	subcmd := p.GetSubcommand()
	{
		type HelpOptions struct {
			SUBCOMMAND string
		}
		shell.R(&shell.SubParserArgs{
			Command:     "help",
			Description: "show help info of a subcoomand",
			Opts:        &HelpOptions{},
			Callback: func(opts *HelpOptions) error {
				s, err := subcmd.SubHelpString(opts.SUBCOMMAND)
				if err != nil {
					return err
				}
				fmt.Print(s)
				return nil
			},
		})
	}
	for _, args := range shell.SubParserArgsMap {
		subcmd.AddSubParser(args.Opts, args.Command, args.Description, args.Callback)
	}
	return p, nil
}

func invokeSubcmd() {
}

func die(err error) {
	fmt.Fprint(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}

func newClientSession() *mcclient.ClientSession {
	client := mcclient.NewClient(
		opts.OsAuthURL,
		30,    // timeout
		false, // debug
		true,  // insecure
		"",    // cert
		"",    //key
	)
	session := client.NewSession(
		context.Background(),
		opts.OsRegionName,
		opts.OsZoneName,
		opts.OsEndpointType,
		nil, // cache token
		"",  // api version
	)
	return session
}

func main() {
	var err error
	p, err := prepareParser()
	if err != nil {
		log.Fatalf("prepare parser: %s", err)
	}
	err = p.ParseArgs(os.Args[1:], false)
	subcmd := p.GetSubcommand()
	subp := subcmd.GetSubParser()
	if err != nil {
		if _, ok := err.(*structarg.NotEnoughArgumentsError); ok && opts.SUBCOMMAND == "" {
			switch {
			case opts.Help:
				fmt.Printf("%s", p.HelpString())
			case opts.Version:
				fmt.Printf("%s\n", version.GetJsonString())
			default:
				// interactive mode
			}
			return
		}
		// print usage
		log.Fatalf("parse args: %s", err)
		if subp != nil {
			fmt.Print(subp.Usage())
		} else {
			fmt.Print(p.Usage())
		}
		die(err)
	}

	// execute single command
	{
		var err error
		subOpts := subp.Options()
		if opts.SUBCOMMAND == "help" {
			err = subcmd.Invoke(subOpts)
		} else {
			s := newClientSession()
			err = subcmd.Invoke(s, subOpts)
		}
		if err != nil {
			die(err)
		}
	}
	return
}
