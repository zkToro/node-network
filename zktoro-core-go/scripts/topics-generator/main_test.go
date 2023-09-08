package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"zktoro/zktoro-core-go/contracts/generated/contract_agent_registry_0_1_4"
	"zktoro/zktoro-core-go/contracts/generated/contract_dispatch_0_1_4"
	"zktoro/zktoro-core-go/contracts/generated/contract_scanner_registry_0_1_3"
)

const evtScannerUpdated = "ScannerUpdated(uint256,uint256,string)"
const evtScannerEnabled = "ScannerEnabled(uint256,bool,uint8,bool)"
const evtAgentUpdated = "AgentUpdated(uint256,address,string,uint256[])"
const evtAgentEnabled = "AgentEnabled(uint256,bool,uint8,bool)"
const evtLink = "Link(uint256,uint256,bool)"

var evtScannerUpdatedTopic = crypto.Keccak256Hash([]byte(evtScannerUpdated)).Hex()
var evtScannerEnabledTopic = crypto.Keccak256Hash([]byte(evtScannerEnabled)).Hex()
var evtAgentUpdatedTopic = crypto.Keccak256Hash([]byte(evtAgentUpdated)).Hex()
var evtAgentEnabledTopic = crypto.Keccak256Hash([]byte(evtAgentEnabled)).Hex()
var evtLinkTopic = crypto.Keccak256Hash([]byte(evtLink)).Hex()

func TestTopicGeneration(t *testing.T) {
	assert.Equal(t, evtAgentEnabledTopic, contract_agent_registry_0_1_4.AgentEnabledTopic)
	assert.Equal(t, evtAgentUpdatedTopic, contract_agent_registry_0_1_4.AgentUpdatedTopic)
	assert.Equal(t, evtScannerUpdatedTopic, contract_scanner_registry_0_1_3.ScannerUpdatedTopic)
	assert.Equal(t, evtScannerEnabledTopic, contract_scanner_registry_0_1_3.ScannerEnabledTopic)
	assert.Equal(t, evtLinkTopic, contract_dispatch_0_1_4.LinkTopic)
}
