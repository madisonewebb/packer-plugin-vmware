# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_version = ">= 1.7.0"
  required_plugins {
    vmware = {
      version = ">= 1.0.7"
      source  = "github.com/hashicorp/vmware"
    }
  }
}

build {
  sources = ["source.vmware-iso.debian"]
  
  // Create SSH directory with proper permissions first
  provisioner "shell" {
    inline = [
      "mkdir -p /home/vagrant/.ssh",
      "chmod 700 /home/vagrant/.ssh"
    ]
  }

  // Add the Vagrant public key to the authorized_keys file
  provisioner "shell" {
    inline = [
      "curl -fsSL https://raw.githubusercontent.com/hashicorp/vagrant/master/keys/vagrant.pub >> /home/vagrant/.ssh/authorized_keys",
      "chown -R vagrant:vagrant /home/vagrant/.ssh"
    ]
  }
  
  // Add your public SSH key to the authorized_keys
  provisioner "file" {
    source      = "/Users/madison/.ssh/id_ed25519.pub"
    destination = "/home/vagrant/.ssh/authorized_keys"
  }

  // Update the system packages + kernel, then install open-vm-tools
  provisioner "shell" {
      inline = [
        "echo 'Updating package lists...'",
        "sudo apt update",
        "echo 'Upgrading packages...'",
        "sudo apt upgrade -y", 
        "sudo apt full-upgrade -y",
        "echo 'Installing open-vm-tools...'",
        "sudo apt install -y open-vm-tools"
      ]
  }
}
