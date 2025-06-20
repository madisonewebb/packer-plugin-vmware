// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepExport represents a step to export a virtual machines to specific formats.
type StepExport struct {
	Format         string
	SkipExport     bool
	VMName         string
	OVFToolOptions []string
	OutputDir      *string
}

func (s *StepExport) generateRemoteExportArgs(c *DriverConfig, displayName string, hidePassword bool, exportOutputPath string) ([]string, error) {

	ovftoolUri := fmt.Sprintf("vi://%s/%s", c.RemoteHost, displayName)
	u, err := url.Parse(ovftoolUri)
	if err != nil {
		return []string{}, err
	}

	password := c.RemotePassword
	if hidePassword {
		password = "<password>"
	}
	u.User = url.UserPassword(c.RemoteUser, password)

	args := []string{
		"--noSSLVerify=true",
		"--skipManifestCheck",
		"-tt=" + s.Format,
		u.String(),
		filepath.Join(exportOutputPath, s.VMName+"."+s.Format),
	}
	return append(s.OVFToolOptions, args...), nil
}

func (s *StepExport) generateLocalExportArgs(exportOutputPath string) ([]string, error) {
	args := []string{
		filepath.Join(exportOutputPath, s.VMName+".vmx"),
		filepath.Join(exportOutputPath, s.VMName+"."+s.Format),
	}
	return append(s.OVFToolOptions, args...), nil
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("driverConfig").(*DriverConfig)
	ui := state.Get("ui").(packersdk.Ui)
	driver := state.Get("driver").(Driver)

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	// load output path from state. If it doesn't exist, just use the local
	// outputdir.
	exportOutputPath, ok := state.Get("export_output_path").(string)
	if !ok || exportOutputPath == "" {
		if *s.OutputDir != "" {
			exportOutputPath = *s.OutputDir
		} else {
			exportOutputPath = s.VMName
		}
	}

	err := os.MkdirAll(exportOutputPath, 0755)
	if err != nil {
		err = fmt.Errorf("error creating export directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Exporting virtual machine...")
	var displayName string
	if v, ok := state.GetOk("display_name"); ok {
		displayName = v.(string)
	}

	var args, uiArgs []string

	ovftool := GetOvfTool()
	if c.RemoteType == "esxi" {
		// Generate arguments for the ovftool command, but obfuscating the
		// password that we can log the command to the UI for debugging.
		uiArgs, err := s.generateRemoteExportArgs(c, displayName, true, exportOutputPath)
		if err != nil {
			err = fmt.Errorf("error generating ovftool export args: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		ui.Sayf("Executing: %s %s", ovftool, strings.Join(uiArgs, " "))
		// Re-run the generate command, this time without obfuscating the
		// password, so we can actually use it.
		args, err = s.generateRemoteExportArgs(c, displayName, false, exportOutputPath)
		if err != nil {
			err = fmt.Errorf("error generating ovftool export args: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		args, err = s.generateLocalExportArgs(exportOutputPath)
		ui.Sayf("Executing: %s %s", ovftool, strings.Join(uiArgs, " "))
	}
	if err != nil {
		err := fmt.Errorf("error generating ovftool export args: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := driver.Export(args); err != nil {
		err = fmt.Errorf("error performing ovftool export: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
