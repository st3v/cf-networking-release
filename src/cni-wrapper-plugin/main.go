package main

import (
	"cni-wrapper-plugin/legacynet"
	"cni-wrapper-plugin/lib"
	"encoding/json"
	"errors"
	"fmt"
	"lib/datastore"
	"lib/filelock"
	"lib/rules"
	"lib/serial"
	"net/http"
	"os"
	"sync"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/020"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/coreos/go-iptables/iptables"
)

func cmdAdd(args *skel.CmdArgs) error {
	n, err := lib.LoadWrapperConfig(args.StdinData)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Get(n.HealthCheckURL)
	if err != nil {
		return fmt.Errorf("could not call health check: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("health check failed with %d", resp.StatusCode))
	}

	pluginController, err := newPluginController(n.IPTablesLockFile)
	if err != nil {
		return err
	}

	result, err := pluginController.DelegateAdd(n.Delegate)
	if err != nil {
		return fmt.Errorf("delegate call: %s", err)
	}

	result020, err := result.GetAsVersion("0.2.0")
	if err != nil {
		return fmt.Errorf("cni delegate plugin result version incompatible: %s", err) // not tested
	}

	containerIP := result020.(*types020.Result).IP4.IP.IP

	if n.RuntimeConfig != nil {
		// Initialize NetOut
		netOutProvider := legacynet.NetOut{
			ChainNamer: &legacynet.ChainNamer{
				MaxLength: 28,
			},
			IPTables:      pluginController.IPTables,
			Converter:     &legacynet.NetOutRuleConverter{},
			GlobalLogging: n.IPTablesASGLogging,
		}
		if err := netOutProvider.Initialize(args.ContainerID, containerIP, n.OverlayNetwork); err != nil {
			return fmt.Errorf("initialize net out: %s", err)
		}

		// Initialize NetIn
		netinProvider := legacynet.NetIn{
			ChainNamer: &legacynet.ChainNamer{
				MaxLength: 28,
			},
			IPTables: pluginController.IPTables,
		}
		err = netinProvider.Initialize(args.ContainerID)

		// Create port mappings
		portMappings := n.RuntimeConfig.PortMappings
		for _, netIn := range portMappings {
			if netIn.HostPort <= 0 {
				return fmt.Errorf("cannot allocate port %d", netIn.HostPort)
			}
			if err := netinProvider.AddRule(args.ContainerID, int(netIn.HostPort), int(netIn.ContainerPort), n.InstanceAddress, containerIP.String()); err != nil {
				return fmt.Errorf("adding netin rule: %s", err)
			}
		}

		// Create egress rules
		netOutRules := n.RuntimeConfig.NetOutRules
		if err := netOutProvider.BulkInsertRules(args.ContainerID, netOutRules, containerIP.String()); err != nil {
			return fmt.Errorf("bulk insert: %s", err) // not tested
		}
	}

	err = pluginController.AddIPMasq(containerIP.String(), n.OverlayNetwork)
	if err != nil {
		return fmt.Errorf("error setting up default ip masq rule: %s", err)
	}

	store := &datastore.Store{
		Serializer: &serial.Serial{},
		Locker: &filelock.Locker{
			Path: n.Datastore,
		},
	}

	var cniAddData struct {
		Metadata map[string]interface{}
	}
	if err := json.Unmarshal(args.StdinData, &cniAddData); err != nil {
		panic(err) // not tested, this should be impossible
	}

	if err := store.Add(args.ContainerID, containerIP.String(), cniAddData.Metadata); err != nil {
		storeErr := fmt.Errorf("store add: %s", err)
		fmt.Fprintf(os.Stderr, "%s", storeErr)
		fmt.Fprintf(os.Stderr, "cleaning up from error")
		err = pluginController.DelIPMasq(containerIP.String(), n.OverlayNetwork)
		if err != nil {
			fmt.Fprintf(os.Stderr, "during cleanup: removing IP masq: %s", err)
		}

		return storeErr
	}

	result030, err := current.NewResultFromResult(result020)
	if err != nil {
		return fmt.Errorf("error converting result to 0.3.0: %s", err) // not tested
	}
	return result030.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	n, err := lib.LoadWrapperConfig(args.StdinData)
	if err != nil {
		return err
	}

	store := &datastore.Store{
		Serializer: &serial.Serial{},
		Locker: &filelock.Locker{
			Path: n.Datastore,
		},
	}

	container, err := store.Delete(args.ContainerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "store delete: %s", err)
	}

	pluginController, err := newPluginController(n.IPTablesLockFile)
	if err != nil {
		return err
	}

	if err := pluginController.DelegateDel(n.Delegate); err != nil {
		fmt.Fprintf(os.Stderr, "delegate delete: %s", err)
	}

	netInProvider := legacynet.NetIn{
		ChainNamer: &legacynet.ChainNamer{
			MaxLength: 28,
		},
		IPTables: pluginController.IPTables,
	}

	if err = netInProvider.Cleanup(args.ContainerID); err != nil {
		fmt.Fprintf(os.Stderr, "net in cleanup: %s", err)
	}

	netOutProvider := legacynet.NetOut{
		ChainNamer: &legacynet.ChainNamer{
			MaxLength: 28,
		},
		IPTables:      pluginController.IPTables,
		Converter:     &legacynet.NetOutRuleConverter{},
		GlobalLogging: n.IPTablesASGLogging,
	}

	if err = netOutProvider.Cleanup(args.ContainerID); err != nil {
		fmt.Fprintf(os.Stderr, "net out cleanup", err)
	}

	err = pluginController.DelIPMasq(container.IP, n.OverlayNetwork)
	if err != nil {
		fmt.Fprintf(os.Stderr, "removing IP masq: %s", err)
	}

	return nil
}

func newPluginController(iptablesLockFile string) (*lib.PluginController, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, err
	}

	iptLocker := &rules.IPTablesLocker{
		FileLocker: &filelock.Locker{Path: iptablesLockFile},
		Mutex:      &sync.Mutex{},
	}
	restorer := &rules.Restorer{}
	lockedIPTables := &rules.LockedIPTables{
		IPTables: ipt,
		Locker:   iptLocker,
		Restorer: restorer,
	}

	pluginController := &lib.PluginController{
		Delegator: lib.NewDelegator(),
		IPTables:  lockedIPTables,
	}
	return pluginController, nil
}

func main() {
	supportedVersions := []string{"0.3.0"}

	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports(supportedVersions...))
}
