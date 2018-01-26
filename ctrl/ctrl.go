package main

import (
	"os"
	"sort"

	_ "IPT/cmd"
	"IPT/cmd/asset"
	"IPT/cmd/bookkeeper"
	. "IPT/cmd/common"
	"IPT/cmd/consensus"
	"IPT/cmd/data"
	"IPT/cmd/debug"
	"IPT/cmd/info"
	"IPT/cmd/multisig"
	"IPT/cmd/privpayload"
	"IPT/cmd/recover"
	"IPT/cmd/smartcontract"
	"IPT/cmd/test"
	"IPT/cmd/wallet"

	"github.com/urfave/cli"
)

var Version string

func main() {
	app := cli.NewApp()
	app.Name = "nodectl"
	app.Version = Version
	app.HelpName = "nodectl"
	app.Usage = "command line tool for IPT blockchain"
	app.UsageText = "nodectl [global options] command [command options] [args]"
	app.HideHelp = false
	app.HideVersion = false
	//global options
	app.Flags = []cli.Flag{
		NewIpFlag(),
		NewPortFlag(),
	}
	//commands
	app.Commands = []cli.Command{
		*consensus.NewCommand(),
		*debug.NewCommand(),
		*info.NewCommand(),
		*test.NewCommand(),
		*wallet.NewCommand(),
		*asset.NewCommand(),
		*privpayload.NewCommand(),
		*data.NewCommand(),
		*bookkeeper.NewCommand(),
		*recover.NewCommand(),
		*multisig.NewCommand(),
		*smartcontract.NewCommand(),
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
