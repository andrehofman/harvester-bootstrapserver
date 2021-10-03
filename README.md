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

### Configuration Data

Ipsum

### How To Build

Ipsum
