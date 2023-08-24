# install-ppa

A simple tool to install a PPA.

## usage

```
install-ppa -help
Usage of install-ppa:
  -apt-config string
        APT Config Path
  -distro string
        Distribution
  -key-id string
        GPG key ID
  -ppa string
        PPA Name
```

## install

```
go install github.com/gjolly/install-ppa/cmd/install-ppa@latest
```

## example

To install a PPA on your system:

```
install-ppa -ppa 'ppa:mozillateam/ppa' -key-id '0AB215679C571D1C8325275B9BDB3D89CE49EC21'
```

To install a PPA in a `chroot`:

```
install-ppa -ppa 'ppa:mozillateam/ppa' -key-id '0AB215679C571D1C8325275B9BDB3D89CE49EC21' \
    -apt-config './chroot/etc/apt' -distro jammy
```

## TODO

 * add support for private PPAs
 * add command to remove a PPA
