# ppa-utils

A simple set of tools to manage PPAs.

## usage

```bash
$ install-ppa -help
Usage of install-ppa:
  -apt-config string
        APT Config Path
  -distro string
        Distribution
  -key-id string
        GPG key ID
  -ppa string
        PPA Name
$ remove-ppa --help
Usage of remove-ppa:
  -apt-config string
        APT Config Path
  -dryrun
        PPA to remove
  -ppa string
        PPA to remove
$ list-ppa --help
Usage of list-ppa:
  -apt-config string
        APT Config Path
  -format string
        Output format (text, json) (default "text")
```

## install

PPA:

```bash
curl -L 'https://keyserver.ubuntu.com/pks/lookup?op=get&search=0xe8de5c81c12b06fe3fc4e35114aaaf80565cd7fb' \
    | gpg --dearmor \
    | sudo tee /etc/apt/keyrings/ppa-utils.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/ppa-utils.gpg] https://ppa.launchpadcontent.net/gjolly/ppa-utils/ubuntu $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/ppa-utils.list
sudo apt update
sudo apt install -y ppa-utils
```

Binaries (only for `amd64`):

```bash
curl -LO https://github.com/gjolly/ppa-utils/releases/latest/download/ppa-utils.tar.gz
```

With Go:

```bash
go install github.com/gjolly/install-ppa/cmd/install-ppa@latest
```

## examples

To install a PPA on your system:

```bash
install-ppa -ppa 'ppa:mozillateam/ppa' -key-id '0AB215679C571D1C8325275B9BDB3D89CE49EC21'
```

To install a PPA in a `chroot`:

```bash
install-ppa -ppa 'ppa:mozillateam/ppa' -key-id '0AB215679C571D1C8325275B9BDB3D89CE49EC21' \
    -apt-config './chroot/etc/apt' -distro jammy
```

## TODO

 * improve `remove-ppa` when removing a line from a file. Make sure it's done nicely
 * add support for private PPAs
 * add command to remove a PPA
 * add tests (especially for `remove-ppa`)
