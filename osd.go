package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/etcd"
	"github.com/portworx/kvdb/mem"

	apiserver "github.com/libopenstorage/openstorage/api/server"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	version = "0.3"
)

func start(c *cli.Context) {
	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
		return
	}

	datastores := []string{mem.Name, etcd.Name}

	// We are in daemon mode.
	file := c.String("file")
	if file == "" {
		fmt.Println("OSD configuration file not specified.  Visit openstorage.org for an example.")
		return
	}
	cfg, err := config.Parse(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	kvdbURL := c.String("kvdb")
	u, err := url.Parse(kvdbURL)
	scheme := u.Scheme
	u.Scheme = "http"

	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil)
	if err != nil {
		fmt.Println("Failed to initialize KVDB: ", u.Scheme, err)
		fmt.Println("Supported datastores: ", datastores)
		return
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		fmt.Println("Failed to initialize KVDB: ", err)
		return
	}

	// Start the cluster state machine, if enabled.
	if cfg.Osd.ClusterConfig.NodeId != "" && cfg.Osd.ClusterConfig.ClusterId != "" {
		_, err = cluster.New(cfg.Osd.ClusterConfig, kv)
		if err != nil {
			fmt.Println("Failed to initialize cluster: ", err)
			return
		}
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		fmt.Println("Starting volume driver: ", d)
		_, err := volume.New(d, v)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", d, err)
			return
		}

		err = apiserver.StartServerAPI(d, 0, config.DriverAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			return
		}

		err = apiserver.StartPluginAPI(d, config.PluginAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume plugin: ", err)
			return
		}

		err = apiserver.StartPluginMgmntAPI(d, config.PluginAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume plugin Mgr : ", err)
			return
		}
	}

	// Daemon does not exit.
	select {}
}

func showVersion(c *cli.Context) {
	fmt.Println("OSD Version:", version)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS:", runtime.GOOS)
	fmt.Println("Arch:", runtime.GOARCH)
}

func main() {
	app := cli.NewApp()
	app.Name = "osd"
	app.Usage = "Open Storage CLI"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json,j",
			Usage: "output in json",
		},
		cli.BoolFlag{
			Name:  osdcli.DaemonAlias,
			Usage: "Start OSD in daemon mode",
		},
		cli.StringSliceFlag{
			Name:  "driver",
			Usage: "driver name and options: name=btrfs,root_vol=/var/openstorage/btrfs",
			Value: new(cli.StringSlice),
		},
		cli.StringFlag{
			Name:  "kvdb,k",
			Usage: "uri to kvdb e.g. kv-mem://localhost, etcd://localhost:4001",
			Value: "kv-mem://localhost",
		},
		cli.StringFlag{
			Name:  "file,f",
			Usage: "file to read the OSD configuration from.",
			Value: "",
		},
	}
	app.Action = start

	app.Commands = []cli.Command{
		{
			Name:        "driver",
			Aliases:     []string{"d"},
			Usage:       "Manage drivers",
			Subcommands: osdcli.DriverCommands(),
		},
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Display version",
			Action:  showVersion,
		},
	}

	for _, v := range drivers {
		if v.driverType == volume.Block {
			bCmds := osdcli.BlockVolumeCommands(v.name)
			clstrCmds := osdcli.ClusterCommands(v.name)
			cmds := append(bCmds, clstrCmds...)
			c := cli.Command{
				Name:        v.name,
				Usage:       fmt.Sprintf("Manage %s storage", v.name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		} else if v.driverType == volume.File {
			fCmds := osdcli.FileVolumeCommands(v.name)
			clstrCmds := osdcli.ClusterCommands(v.name)
			cmds := append(fCmds, clstrCmds...)
			c := cli.Command{
				Name:        v.name,
				Usage:       fmt.Sprintf("Manage %s volumes", v.name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		} else {
			fmt.Println("Unable to start volume plugin: ", errors.New("Unknown driver type."))
			return
		}
	}
	app.Run(os.Args)
}
