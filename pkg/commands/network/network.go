package network

import (
	"github.com/mrtc0/bouheki/pkg/bpf"
	"github.com/mrtc0/bouheki/pkg/config"
	log "github.com/mrtc0/bouheki/pkg/log"
	"github.com/urfave/cli/v2"
)

func loadBytecode(mode string) ([]byte, string, error) {
	bytecode, err := bpf.EmbedFS.ReadFile("bytecode/restricted-network.bpf.o")
	if err != nil {
		return nil, "", err
	}
	return bytecode, "restricted-network", nil
}

func Run(ctx *cli.Context) error {
	path := ctx.String("config")
	conf, err := config.NewConfig(path)
	if err != nil {
		return err
	}

	log.SetFormatter(conf.Log.Format)
	log.SetOutput(conf.Log.Output)
	log.SetRotation(conf.Log.Output, conf.Log.MaxSize, conf.Log.MaxAge)

	bytecode, objName, err := loadBytecode(conf.Network.Mode)
	if err != nil {
		return err
	}

	RunAudit(bytecode, objName, conf)

	return nil
}
