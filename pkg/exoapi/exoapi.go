package exoapi

import (
	"errors"
	"fmt"

	meta "github.com/exoip-k8s/pkg/metadata"
	"github.com/pyr/egoscale/src/egoscale"
)

func hasSecurityGroup(vm *egoscale.VirtualMachine, sgname string) bool {

	for _, sg := range vm.SecurityGroups {
		if sg.Name == sgname {
			return true
		}
	}
	return false
}

// GetSecurityGroupPeers
func GetSecurityGroupPeers(ego *egoscale.Client, sgname string) ([]string, error) {

	peers := make([]string, 0)
	vms, err := ego.ListVirtualMachines()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if hasSecurityGroup(vm, sgname) {
			primaryIP := vm.Nic[0].Ipaddress
			peers = append(peers, fmt.Sprintf("%s", primaryIP))
		}
	}

	return peers, nil
}

// FetchMyNic
func FetchMyNic(ego *egoscale.Client, mserver string) (string, error) {

	instanceID, err := meta.FetchMetadata(mserver, "/latest/instance-id")
	if err != nil {
		return "", err
	}
	vmInfo, err := ego.GetVirtualMachine(instanceID)
	if err != nil {
		return "", err
	}
	if len(vmInfo.Nic) < 1 {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return vmInfo.Nic[0].Id, nil
}

// FindPeerNic
func FindPeerNic(ego *egoscale.Client, ip string) (string, error) {

	vms, err := ego.ListVirtualMachines()
	if err != nil {
		return "", err
	}

	for _, vm := range vms {

		if vm.Nic[0].Ipaddress == ip {
			return vm.Nic[0].Id, nil
		}
	}

	return "", fmt.Errorf("cannot find nic")
}
