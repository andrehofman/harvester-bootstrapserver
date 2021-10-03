# HTTP iPXE Server

This iPXE server is capable of sending specific ipxe scripts, and harvester configuration files (config-create.yaml & config-join.yaml) to the requesting server, based on a MAC-address.  

It uses one configuration file containing the MAC-addresses and plant/region specific configuration.  

The server serves ipxe-scripts based on MAC-addresses, and those ipxe-scripts contain the correct config-create.yaml or config-join.yaml links for that specific server.  

This way it could be used as a bootstrap service for Harvester.


## Requirements

- iPXE script (see [example](#ipxe-script))
- data.yaml   (see [data.yaml](#configuration-data))
- The binary from this code (check [how to build](#how-to-build))

## How it works

lorum ipsum

### iPXE script

Example of an ipxe-script that chainloads the specific ipxe-script:

    #!ipxe
    :loop
    set net0/ip 192.168.178.252
    set net0/netmask 255.255.255.0
    set net0/gateway 192.168.178.1
    ifopen
    iflinkwait --timeout 3000 net0
    ifstat
    chain http://192.168.178.7:10000/ipxe/${net0/hwaddr} || goto loop

What happens with the above script is that it sets a label with `:loop`, Then it sets an ipaddress, netmask and gateway for a specific interface. This depends on your setup, i.e. if you've got an interface that is let's say connected to the management vlan. Then the scripts brings the interface up with `ifopen`, waits for it to be up for 3000 milliseconds with `iflinkwait`. Prints the interface stats with `ifstat`, and proceeds by looking for another ipxe-script with `chain`. In that line the `${net0/hwaddr}` is replaced with the mac-address from that interface. If it fails it returns to the line with `:loop` and does it all again indefinitely.

#### HowTo

- Clone the git repository: git.ipxe.org/ipxe.git, and move into the ipxe directory:

        git clone git.ipxe.org/ipxe.git
        cd ipxe

- Create a file with the ipxe-script, i.e. example.ipxe (adjust where necesarry)
- Build your own ipxe.iso to boot from, execute:

        cd src
        make bin/ipxe.iso EMBED=../example.ipxe

  *Assuming you placed the example.ipxe in the directory where you cloned the ipxe git-repository.*

  If everything went well, the file `bin/ipxe.iso` has now the example.ipxe embeded, and when booting from that iso-file it will be executed by default.

Source: [https://ipxe.org/embed](https://ipxe.org/embed)

As I use Libvirtd (KVM) I set the boot order to `hd,cdrom` that way the machine will try to boot from disk, which it can't, wand proceeds to boot from the cdrom - the ipxe.iso file. Fetches the specific ipxe script for that machine, and continues with the installation. Upon the next boot the `hd` is bootable so it will boot from the fresh installed Harvester.

### Configuration Data

Imagine the following:

- There are 2 locations, that have both an isolated Harvester Cluster setup. Let's call the `location_X` & `location_Y`.
- Both locations have their own specific configuration for about IP addresses, and maybe even network interfaces due to differences of hardware.

The configuration file consists of 2 main parts - `config` & `maclist`. Both are a list of objects, confguration objects, and a list of macaddresses that is used to provide each machine its specific ipxe-script and configuration.  
`config` contains a list of configuration

For more information regarding the specific configuration parameters for Harvester see: [https://github.com/harvester/docs/blob/main/docs/install/harvester-configuration.md](https://github.com/harvester/docs/blob/main/docs/install/harvester-configuration.md)

An example configuration file showing different configuration for each location:

```yaml
config:
  - cluster: location_X
    config_create: &location_X_create
      token: d506621a3d4acdf60d62475f1b2cb681
      os:
        hostname: harvester01
        ssh_authorized_keys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII9lnqlqqlHszXPP8zHFlqrQ4utzVJMSTJI2Qba+zE1s sshkey_X
          - github:your_username
        password: 5up3rS3cr3t
        dns_nameservers:
          - 192.168.178.3
        ntp_servers:
          - 0.pool.ntp.org
          - 1.pool.ntp.org
      install:
        mode: create
        networks:
          harvester-mgmt:
            interfaces:
            - name: enp1s0
            method: dhcp
        device: /dev/vda
        iso_url: http://192.168.178.7/harvester/0.3.0/harvester-amd64.iso
        vipMode: static
        vip: 192.168.178.40
        poweroff: false
    config_join: &location_X_join
      server_url: https://192.168.178.41:8443
      token: d506621a3d4acdf60d62475f1b2cb681
      os:
        ssh_authorized_keys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII9lnqlqqlHszXPP8zHFlqrQ4utzVJMSTJI2Qba+zE1s sshkey_X
          - github:your_username
        password: 5up3rS3cr3t 
        dns_nameservers:
          - 192.168.178.3
        ntp_servers:
          - 0.pool.ntp.org
          - 1.pool.ntp.org
      install:
        mode: join
        networks:
          harvester-mgmt:
            interfaces:
            - name: enp1s0
            method: dhcp
        mgmt_interface: enp1s0
        device: /dev/vda
        iso_url: http://192.168.178.7/harvester/0.3.0/harvester-amd64.iso
        poweroff: false
  - cluster: location_Y
    config_create: &loaction_Y_join
      token: 612c99d263fe6b8acb56f0b964e07d0a
      os:
        hostname: harvester01
        ssh_authorized_keys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII9lnqlqqlHszXPP8zHFlqrQ4utzVJMSTJI2Qba+zE1s andre@amd
          - github:your_username
        password: 5up3rS3cr3t
        dns_nameservers:
          - 172.16.128.3
        ntp_servers:
          - ntp.your.org
          - ntp.your.org
      install:
        mode: create
        networks:
          harvester-mgmt:
            interfaces:
            - name: enp1s0
            method: dhcp
        device: /dev/vda
        iso_url: http://192.168.178.7/harvester/0.3.0/harvester-amd64.iso
        vipMode: static
        vip: 172.16.128.40
        poweroff: false
    config_join: &location_Y_join
      server_url: https://172.16.128.41:8443
      token: 612c99d263fe6b8acb56f0b964e07d0a
      os:
        ssh_authorized_keys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII9lnqlqqlHszXPP8zHFlqrQ4utzVJMSTJI2Qba+zE1s andre@amd
          - github:your_username
        password: 5up3rS3cr3t 
        dns_nameservers:
          - 172.16.128.3
        ntp_servers:
          - 0.pool.ntp.org
          - 1.pool.ntp.org
      install:
        mode: join
        networks:
          harvester-mgmt:
            interfaces:
            - name: enp1s0
            method: dhcp
        mgmt_interface: enp1s0
        device: /dev/vda
        iso_url: http://192.168.178.7/harvester/0.3.0/harvester-amd64.iso
        poweroff: false
maclist:
# harvester v0.2.0
- macaddress: 52:54:00:52:cb:c1
  values:
    ipaddress: 192.168.178.31
    netmask: 255.255.255.0
    gateway: 192.168.178.1
    interface: enp1s0
    version: "0.2.0"
    type: create
    config: 
      << : *location_X_create
- macaddress: 52:54:00:7d:1a:a2
  values:
    ipaddress: 192.168.178.32
    netmask: 255.255.255.0
    gateway: 192.168.178.1
    interface: enp1s0
    version: "0.2.0"
    type: join
    config:
      << : *location_X_join
- macaddress: 52:54:00:31:3e:b8
  values:
    ipaddress: 192.168.178.33
    netmask: 255.255.255.0
    gateway: 192.168.178.1
    interface: enp1s0
    version: "0.2.0"
    type: join
    config:
      << : *location_X_join
# harvester v0.3.0
- macaddress: 52:54:00:c1:6b:56
  values:
    ipaddress: 172.16.128.41
    netmask: 255.255.252.0
    gateway: 172.16.128.254
    interface: ens3
    version: 0.3.0
    type: create
    config:
      << : *location_Y_create
- macaddress: 52:54:00:14:26:90
  values:
    ipaddress: 172.16.128.42
    netmask: 255.255.252.0
    gateway: 172.16.128.254
    interface: ens3
    version: 0.3.0
    type: join
    config:
      << : *location_Y_join
- macaddress: 52:54:00:27:8c:78
  values:
    ipaddress: 172.16.128.43
    netmask: 255.255.252.0
    gateway: 172.16.128.254
    interface: ens3
    version: 0.3.0
    type: join
    config:
      << : *location_Y_join
- macaddress: 52:54:00:9d:cb:2f
  values:
    ipaddress: 172.16.128.44
    netmask: 255.255.252.0
    gateway: 172.16.128.254
    interface: ens3
    version: 0.3.0
    type: join
    config:
      << : *location_Y_join

```

### How To Build

Ipsum
