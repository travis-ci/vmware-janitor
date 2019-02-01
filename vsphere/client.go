package vsphere

import (
	"context"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/travis-ci/jupiter-brain/pkg/vsphereutil"
	vspherejanitor "github.com/travis-ci/vsphere-janitor"
	"github.com/travis-ci/vsphere-janitor/log"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Client struct {
	clientProvider vsphereutil.ClientProvider
}

func NewClient(ctx context.Context, u *url.URL, insecure bool) (*Client, error) {
	return &Client{
		clientProvider: vsphereutil.NewClientProvider(u, insecure),
	}, nil
}

func (c *Client) ListVMs(ctx context.Context, path string) ([]vspherejanitor.VirtualMachine, error) {
	folder, err := c.folder(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "error finding folder")
	}

	rawVMs, err := folder.Children(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error listing contents of VM folder")
	}

	vms := make([]vspherejanitor.VirtualMachine, 0, len(rawVMs))

	for _, rawVM := range rawVMs {
		ovm, ok := rawVM.(*object.VirtualMachine)
		if !ok {
			continue
		}

		mvm := &mo.VirtualMachine{}

		err = ovm.Properties(ctx, ovm.Reference(), []string{"config", "summary"}, mvm)
		if err != nil {
			log.WithContext(ctx).WithError(err).Info("couldn't get properties for VM")
			continue
		}

		vm := &VirtualMachine{
			vm:  ovm,
			mvm: mvm,
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

func (c *Client) folder(ctx context.Context, path string) (*object.Folder, error) {
	client, err := c.clientProvider.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get govmomi client")
	}

	searchIndex := object.NewSearchIndex(client.Client)

	folderRef, err := searchIndex.FindByInventoryPath(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "error looking for VM folder")
	}

	if folderRef == nil {
		return nil, errors.New("couldn't find VM folder")
	}

	folder, ok := folderRef.(*object.Folder)
	if !ok {
		return nil, errors.Errorf("VM folder is not a folder but a %T", folderRef)
	}

	return folder, nil
}

type VirtualMachine struct {
	vm  *object.VirtualMachine
	mvm *mo.VirtualMachine
}

func (vm *VirtualMachine) Name() string {
	if vm.mvm.Config == nil {
		return "<unnamed>"
	}

	return vm.mvm.Config.Name
}

func (vm *VirtualMachine) ID() string {
	if vm.mvm.Config == nil {
		return ""
	}

	return vm.mvm.Config.Uuid
}

func (vm *VirtualMachine) Uptime() time.Duration {
	return time.Duration(vm.mvm.Summary.QuickStats.UptimeSeconds) * time.Second
}

func (vm *VirtualMachine) BootTime() *time.Time {
	return vm.mvm.Summary.Runtime.BootTime
}

func (vm *VirtualMachine) PoweredOn() bool {
	return vm.mvm.Summary.Runtime.PowerState == types.VirtualMachinePowerStatePoweredOn
}

func (vm *VirtualMachine) PowerOff(ctx context.Context) error {
	task, err := vm.vm.PowerOff(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't create power off task")
	}

	err = task.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't power off instance")
	}

	return nil
}

func (vm *VirtualMachine) Destroy(ctx context.Context) error {
	task, err := vm.vm.Destroy(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't create destroy task")
	}

	err = task.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't destroy instance")
	}

	return nil
}
