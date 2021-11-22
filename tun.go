package litetun

import (
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"unsafe"

	"github.com/vishvananda/netlink"
)

func NewTun(name string, ipCIDR *string) (tun *Tun, err error) {

	t := &Tun{
		name: name,
	}

	if ipCIDR != nil {
		err := t.SetAddr(*ipCIDR)
		if err != nil {
			return nil, err
		}
	}

	return t, nil

}

type Tun struct {
	ip net.IP
	network *net.IPNet
	name string
	fd int
	link netlink.Link
	isOpen bool
}

func (t *Tun) SetName(name string) {

    t.name = name

}

func (t *Tun) Name() string {

    return t.name

}

func (t *Tun) SetAddr(ipCIDR string) error {
	ip, ipnet, err := net.ParseCIDR(ipCIDR)
	if err != nil {
		return err
	}

    t.ip = ip
	t.network = ipnet

	return t.setIP()

}

func (t *Tun) SetIP(ip net.IP) error {

    t.ip = ip
	return t.setIP()

}

func (t *Tun) IP() net.IP {

    return t.ip

}

func (t *Tun) SetNetwork(ipnet *net.IPNet) error {

    t.network = ipnet
	return t.setIP()

}

func (t *Tun) Network() *net.IPNet {

    return t.network

}

func (t *Tun) Read(b []byte) (n int, err error) {

    return unix.Read(t.fd, b)

}

func (t *Tun) Write(b []byte) (n int, err error) {

    return unix.Write(t.fd, b)

}

func (t *Tun) Open() error {

    if t.isOpen {
        return errors.New("tunnel is already open")
    }

	return t.open(unix.IFF_TUN|unix.IFF_NO_PI, false)

}

func (t *Tun) IsOpen() bool {

    return t.isOpen

}

func (t *Tun) Close() error {

	if t.isOpen {
		return unix.Close(t.fd)
	}
	return nil

}

func (t *Tun) Up() error {

	if t.link == nil {
		if err := t.findLink(); err != nil {
            return err
        }
	}
	return netlink.LinkSetUp(t.link)

}

func (t *Tun) Down() error {

	if t.link == nil {
		if err := t.findLink(); err != nil {
			return err
		}
	}
	return netlink.LinkSetDown(t.link)

}

func (t *Tun) open(flags uint16, nonblocking bool) error {

	// Open the tun device
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR, 0)
	if err != nil {
        return err
    }

	// Create the interface request
	var ifr Ifr
	copy(ifr.Name[:], t.name)
	ifr.Flags = flags

	// Send IOCTL to create the device
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&ifr)))
	if errno != 0 {
		unix.Close(fd)
        return errno
    }

	// Set non-blocking
	if err = unix.SetNonblock(fd, nonblocking); err != nil {
		unix.Close(fd)
		return err
	}

	t.fd = fd
	t.isOpen = true
	return nil

}

func (t *Tun) setIP() error {

	// Get interface
	if t.link == nil {
		if err := t.findLink(); err != nil {
            return err
        }
	}

	// Configure the ip address
	n := t.network
	n.IP = t.ip
	err := netlink.AddrAdd(t.link, &netlink.Addr{IPNet: n})
	if err != nil {
		return err
	}

	// Bring the link up
	return t.Up()

}

func (t *Tun) findLink() (err error) {

	t.link, err = netlink.LinkByName(t.name)
	if err != nil {
        return err
    }

	return nil
}