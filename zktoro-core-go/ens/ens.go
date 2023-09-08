package ens

import (
	"bytes"
	"fmt"

	"zktoro/zktoro-core-go/domain/registry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/wealdtech/go-ens/v3"
)

const (
	DispatchContract            = "dispatch.zktoro.eth"
	AgentRegistryContract       = "agents.registries.zktoro.eth"
	ScannerRegistryContract     = "scanners.registries.zktoro.eth"
	ScannerPoolRegistryContract = "pools.registries.zktoro.eth"
	ScannerNodeVersionContract  = "scanner-node-version.zktoro.eth"
	StakingContract             = "staking.zktoro.eth"
	ZktoroContract              = "zktoro.eth"
	MigrationContract           = "migration.zktoro.eth"
	RewardsContract             = "rewards.zktoro.eth"
	StakeAllocatorContract      = "allocator.zktoro.eth"
)

// ENS resolves inputs.
type ENS interface {
	Resolver
	ResolveRegistryContracts() (*registry.RegistryContracts, error)
}

// ENSStore wraps the ENS client which interacts with namespace contract(s).
type ENSStore struct {
	Resolver
}

// NewENSStore creates a new store.
func NewENSStore(backend bind.ContractBackend) *ENSStore {
	return &ENSStore{Resolver: &ENSResolver{backend: backend}}
}

// DialENSStore dials an Ethereum API and creates a new store.
func DialENSStore(rpcUrl string) (*ENSStore, error) {
	client, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	return &ENSStore{Resolver: &ENSResolver{backend: ethclient.NewClient(client)}}, nil
}

// DialENSStoreAt dials an Ethereum API and creates a new store that works with a resolver at given address.
func DialENSStoreAt(rpcUrl, resolverAddr string) (*ENSStore, error) {
	client, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	return &ENSStore{Resolver: &ENSResolver{backend: ethclient.NewClient(client), resolverAddr: resolverAddr}}, nil
}

// NewENStoreWithResolver creates a new store with custom resolver.
func NewENStoreWithResolver(resolver Resolver) *ENSStore {
	return &ENSStore{Resolver: resolver}
}

// Resolver resolves inputs.
type Resolver interface {
	Resolve(input string) (common.Address, error)
}

// ENSResolver resolves names from an ENS contract.
type ENSResolver struct {
	backend      bind.ContractBackend
	resolverAddr string
}

// Resolve resolves an input to an address.
func (ensResolver *ENSResolver) Resolve(input string) (common.Address, error) {
	if len(ensResolver.resolverAddr) == 0 {
		return ens.Resolve(ensResolver.backend, input)
	}
	resolver, err := ens.NewResolverAt(ensResolver.backend, input, common.HexToAddress(ensResolver.resolverAddr))
	if err != nil {
		return common.Address{}, err
	}
	// Resolve the domain
	address, err := resolver.Address()
	if err != nil {
		return ens.UnknownAddress, err
	}
	if bytes.Equal(address.Bytes(), ens.UnknownAddress.Bytes()) {
		return ens.UnknownAddress, fmt.Errorf("no address for %s", input)
	}
	return address, nil
}

// ResolverFunc helps implementing a custom resolver with a function.
type ResolverFunc func(input string) (common.Address, error)

// Resolve implements the Resolver interface.
func (rf ResolverFunc) Resolve(input string) (common.Address, error) {
	return rf(input)
}

func (ensstore *ENSStore) ResolveRegistryContracts() (*registry.RegistryContracts, error) {
	agentReg, err := ensstore.Resolve(AgentRegistryContract)
	if err != nil {
		return nil, err
	}

	scannerReg, err := ensstore.Resolve(ScannerRegistryContract)
	if err != nil {
		return nil, err
	}

	scannerPoolReg, err := ensstore.Resolve(ScannerPoolRegistryContract)
	if err != nil {
		return nil, err
	}

	dispatch, err := ensstore.Resolve(DispatchContract)
	if err != nil {
		return nil, err
	}

	scannerNodeVersion, err := ensstore.Resolve(ScannerNodeVersionContract)
	if err != nil {
		return nil, err
	}

	zktoroStaking, err := ensstore.Resolve(StakingContract)
	if err != nil {
		return nil, err
	}

	zktoro, err := ensstore.Resolve(ZktoroContract)
	if err != nil {
		return nil, err
	}

	migration, err := ensstore.Resolve(MigrationContract)
	if err != nil {
		return nil, err
	}

	rewards, err := ensstore.Resolve(RewardsContract)
	if err != nil {
		return nil, err
	}

	allocator, err := ensstore.Resolve(StakeAllocatorContract)
	if err != nil {
		return nil, err
	}

	regContracts := &registry.RegistryContracts{
		AgentRegistry:       agentReg,
		ScannerRegistry:     scannerReg,
		ScannerPoolRegistry: scannerPoolReg,
		Dispatch:            dispatch,
		ScannerNodeVersion:  scannerNodeVersion,
		ZktoroStaking:       zktoroStaking,
		Zktoro:              zktoro,
		Migration:           migration,
		Rewards:             rewards,
		StakeAllocator:      allocator,
	}

	return regContracts, nil

}
