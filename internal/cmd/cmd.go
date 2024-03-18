package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/abxuz/go-vhostd/internal/service"
	_ "github.com/abxuz/go-vhostd/internal/service/logic"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var (
		config string
		init   bool
	)

	c := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		Args: cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			zone, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			time.Local = zone

			service.Cfg.SetFilePath(config, init)
			service.Proxy.Init()
			service.Api.Init()

			cfg, err := service.Cfg.LoadFromFile()
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			service.Cfg.SaveToMemory(cfg)

			func() {
				service.Cfg.MemoryLock(true)
				defer service.Cfg.MemoryUnlock(true)
				service.Proxy.Reload(cfg)
				service.Api.Reload(cfg)
			}()

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs
		},
	}

	c.Flags().StringVarP(&config, "config", "c", "config.yaml", "config file path")
	c.Flags().BoolVarP(&init, "init", "i", false, "auto initialize config file")
	c.MarkFlagFilename("config")
	return c
}
