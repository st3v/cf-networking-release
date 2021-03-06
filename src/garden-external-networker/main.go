package main

import (
	"flag"
	"fmt"
	"garden-external-networker/bindmount"
	"garden-external-networker/cni"
	"garden-external-networker/config"
	"garden-external-networker/ipc"
	"garden-external-networker/manager"
	"garden-external-networker/port_allocator"
	"lib/filelock"
	"lib/serial"
	"os"

	"code.cloudfoundry.org/lager"
)

var (
	action string
	handle string
	cfg    config.Config
)

func parseArgs(allArgs []string) error {
	var configFilePath string

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	flagSet.StringVar(&action, "action", "", "")
	flagSet.StringVar(&handle, "handle", "", "")
	flagSet.StringVar(&configFilePath, "configFile", "", "")

	err := flagSet.Parse(allArgs[1:])
	if err != nil {
		return err
	}
	if len(flagSet.Args()) > 0 {
		return fmt.Errorf("unexpected extra args: %+v", flagSet.Args())
	}

	if handle == "" {
		return fmt.Errorf("missing required flag 'handle'")
	}

	if configFilePath == "" {
		return fmt.Errorf("missing required flag 'configFile'")
	}

	cfg, err = config.New(configFilePath)
	if err != nil {
		return err
	}

	if action == "" {
		return fmt.Errorf("missing required flag 'action'")
	}

	return nil
}

func main() {
	logger := lager.NewLogger("container-networking.garden-external-networker")
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.INFO))
	if err := mainWithError(logger); err != nil {
		logger.Error("error", err)
		os.Exit(1)
	}
}

func mainWithError(logger lager.Logger) error {
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		return fmt.Errorf("this is a plugin for Garden-runC.  Don't run it directly.")
	}

	err := parseArgs(os.Args)
	if err != nil {
		return fmt.Errorf("parse args: %s", err)
	}
	logger.Info("action", lager.Data{"action": action})

	cniLoader := &cni.CNILoader{
		PluginDir: cfg.CniPluginDir,
		ConfigDir: cfg.CniConfigDir,
		Logger:    logger,
	}

	networks, err := cniLoader.GetNetworkConfigs()
	if err != nil {
		return fmt.Errorf("load cni config: %s", err)
	}

	cniController := &cni.CNIController{
		Logger:         logger,
		CNIConfig:      cniLoader.GetCNIConfig(),
		NetworkConfigs: networks,
	}

	mounter := &bindmount.Mounter{}

	locker := &filelock.Locker{Path: cfg.StateFilePath}
	tracker := &port_allocator.Tracker{
		Logger:    logger,
		StartPort: cfg.StartPort,
		Capacity:  cfg.TotalPorts,
	}
	serializer := &serial.Serial{}
	portAllocator := &port_allocator.PortAllocator{
		Tracker:    tracker,
		Serializer: serializer,
		Locker:     locker,
	}

	manager := &manager.Manager{
		Logger:        logger,
		CNIController: cniController,
		Mounter:       mounter,
		BindMountRoot: cfg.BindMountDir,
		PortAllocator: portAllocator,
	}

	mux := ipc.Mux{
		Up:   manager.Up,
		Down: manager.Down,
	}

	return mux.Handle(action, handle, os.Stdin, os.Stdout)
}
