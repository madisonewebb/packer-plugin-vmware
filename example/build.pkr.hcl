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
  
  // Add your public SSH key to the authorized_keys
  provisioner "file" {
    source      = "/Users/madison/.ssh/id_ed25519.pub"
    destination = "/home/vagrant/.ssh/authorized_keys"
  }
  
  // Set proper permissions on the authorized_keys file
  provisioner "shell" {
    inline = [
      "chmod 600 /home/vagrant/.ssh/authorized_keys",
      "chown -R vagrant:vagrant /home/vagrant/.ssh"
    ]
  }

  // Update the system packages
  provisioner "shell" {
      inline = ["sudo apt-get update", "sudo apt-get upgrade -y"]
  }
}
