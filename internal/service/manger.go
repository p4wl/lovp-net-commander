package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"

	"net-commander-server/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
todo:
given a slice of subnets
- remove arbitrary subnet
- new created subnet should take first free spot

given a subnet
- reserve ip addr
- free ip addr
- new reserve ip addr should take first free spot
*/

type NetManager struct {
	logger *slog.Logger
	repo   repository.Queries
	ctx    context.Context
}

var (
	mask         string = "27"
	startingAddr string = fmt.Sprintf("10.10.0.0/%s", mask)
	ipBlockSize  int    = 32 // mask /27 produce 32 ip addr per net
)

func createSubnets(n int) ([]*net.IPNet, error) {
	_, baseNet, err := net.ParseCIDR(startingAddr)
	if err != nil {
		return nil, err
	}

	baseIP := baseNet.IP.To4()

	subnets := make([]*net.IPNet, 0, n)
	for i := 0; i < n; i++ {
		nextNetIP := make(net.IP, len(baseIP))
		copy(nextNetIP, baseIP)

		// first two octets are out of bounds - by design only 10.10.x.x/27 networks
		// nextNetIP[0]
		// nextNetIP[1]
		nextNetIP[3] += byte((i * ipBlockSize) % 256)
		nextNetIP[2] += byte((i * ipBlockSize) / 256)

		_, nextSubnet, err := net.ParseCIDR(fmt.Sprintf("%s/%s", nextNetIP.String(), mask))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		subnets = append(subnets, nextSubnet)
	}

	return subnets, nil
}

func nextIP(ip net.IP, network *net.IPNet) (net.IP, error) {
	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("invalid IPv4 address")
	}

	broadcast := make(net.IP, len(ip))
	for i := 0; i < len(ip); i++ {
		broadcast[i] = network.IP[i] | ^network.Mask[i]
	}

	// If current IP is broadcast, return error
	if ip.Equal(broadcast) {
		return nil, errors.New("reached broadcast address")
	}

	// Copy to avoid mutating input
	next := make(net.IP, len(ip))
	copy(next, ip)

	// Increment IP
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}

	// Verify it’s still within the network
	if !network.Contains(next) {
		return nil, errors.New("next IP outside of subnet")
	}

	return next, nil
}

func broadcast(n *net.IPNet) net.IP {
	ip := n.IP.To4()
	mask := n.Mask

	bcast := make(net.IP, len(ip))
	for i := 0; i < len(ip); i++ {
		bcast[i] = ip[i] | ^mask[i]
	}
	return bcast
}

func firstIP(network *net.IPNet) net.IP {
	ip := network.IP.To4()
	firstIP := make(net.IP, len(ip))
	copy(firstIP, ip)
	firstIP[3] += 1
	return firstIP
}

func lastIP(network *net.IPNet) net.IP {
	// calc broadcast address and return addr-1
	broadcast := broadcast(network)
	broadcast[3] -= 1

	return broadcast
}

// ############### old utils end ################
/*
	calculate mask for a given number of hosts
*/
func MaskForHosts(N int) (net.IPMask, error) {
	if N <= 0 {
		return nil, fmt.Errorf("number of hosts must be positive")
	}

	// Find the smallest power of two that fits N + 2 (network + broadcast)
	needed := N + 2
	bits := 32
	hostBits := 0

	for (1 << hostBits) < needed {
		hostBits++
	}

	if hostBits > bits {
		return nil, fmt.Errorf("too many hosts for IPv4: %d", N)
	}

	prefix := bits - hostBits
	mask := net.CIDRMask(prefix, bits)
	return mask, nil
}

/*
sIP - existing subnet IP
mask - mask

return: IP of next subnet (ex. 192.168.1.0, /27) -> 192.168.1.32
*/
func NextSubnet(sIP *net.IP, mask *net.IPMask) (*net.IP, error) {
	if sIP == nil || mask == nil {
		return nil, fmt.Errorf("invalid arguments: sIP or mask is nil")
	}

	ip := *sIP
	ipMask := *mask

	ip16 := ip.To16()
	if ip16 == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	ipInt := big.NewInt(0).SetBytes(ip16)

	ones, bits := ipMask.Size()
	subnetSize := big.NewInt(1)
	subnetSize.Lsh(subnetSize, uint(bits-ones))

	nextIPInt := big.NewInt(0).Add(ipInt, subnetSize)

	// Check overflow
	if nextIPInt.BitLen() > bits {
		return nil, fmt.Errorf("overflow: next subnet exceeds address space")
	}

	nextIPBytes := nextIPInt.Bytes()

	if len(nextIPBytes) < net.IPv6len {
		padded := make([]byte, net.IPv6len)
		copy(padded[net.IPv6len-len(nextIPBytes):], nextIPBytes)
		nextIPBytes = padded
	}

	next := net.IP(nextIPBytes)
	if ipv4 := next.To4(); ipv4 != nil {
		next = ipv4
	}

	return &next, nil
}

func NewNetManger(db *pgxpool.Pool, ctx context.Context, logger *slog.Logger) *NetManager {
	return &NetManager{
		logger,
		*repository.New(db),
		ctx,
	}
}

func (m *NetManager) RegisterUser(username string, email string) error {
	user, err := m.repo.CreateUser(m.ctx, repository.CreateUserParams{
		Username: username,
		Email:    email,
	})
	if err != nil {
		return err
	}

	m.logger.Info(fmt.Sprintf("New user created with id: %d, username: %s", user.ID, user.Username))
	return nil
}

func (m *NetManager) checkUserNetwork(userId int32) (bool, error) {
	has, err := m.repo.OwnerHasNetwork(m.ctx, pgtype.Int4{
		Int32: userId,
		Valid: true,
	})
	if err != nil {
		return false, err
	}

	return has, nil
}

func (m *NetManager) CreateNetwork(networkName string, owner string, hosts int) (*repository.Network, error) {
	userId := int32(0) // todo: rework function to use arguments
	has, err := m.checkUserNetwork(userId)
	if err != nil {
		return nil, err
	}

	if has {
		return nil, fmt.Errorf("user %d already has a network", userId)
	}

	nets, err := m.repo.ListNetworks(m.ctx)
	if err != nil {
		return nil, err
	}

	// todo: assume networks are not sorted
	subnetIP := nets[len(nets)-1].Cidr.IP

	mask, err := MaskForHosts(10) // todo: consider dynamic number of hosts
	if err != nil {
		return nil, err
	}

	newSubnet, err := NextSubnet(&subnetIP, &mask)
	if err != nil {
		return nil, err
	}

	n, err := m.repo.CreateNetwork(m.ctx, repository.CreateNetworkParams{
		Name: "",
		Cidr: net.IPNet{
			IP:   *newSubnet,
			Mask: mask,
		},
		OwnerID: pgtype.Int4{
			Int32: userId,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	m.logger.Info(fmt.Sprintf("New network created. CIDR: %s for User: %d.", &n.Cidr, userId))

	return &n, nil
}
