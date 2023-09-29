// Code generated by MockGen. DO NOT EDIT.
// Source: services/components/botio/bot_client.go

// Package mock_botio is a generated GoMock package.
package mock_botio

import (
	reflect "reflect"

	domain "zktoro/zktoro-core-go/domain"
	protocol "zktoro/zktoro-core-go/protocol"
	config "zktoro/config"
	botreq "zktoro/services/components/botio/botreq"
	gomock "github.com/golang/mock/gomock"
)

// MockBotClient is a mock of BotClient interface.
type MockBotClient struct {
	ctrl     *gomock.Controller
	recorder *MockBotClientMockRecorder
}

// MockBotClientMockRecorder is the mock recorder for MockBotClient.
type MockBotClientMockRecorder struct {
	mock *MockBotClient
}

// NewMockBotClient creates a new mock instance.
func NewMockBotClient(ctrl *gomock.Controller) *MockBotClient {
	mock := &MockBotClient{ctrl: ctrl}
	mock.recorder = &MockBotClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBotClient) EXPECT() *MockBotClientMockRecorder {
	return m.recorder
}

// BlockRequestCh mocks base method.
func (m *MockBotClient) BlockRequestCh() chan<- *botreq.BlockRequest {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockRequestCh")
	ret0, _ := ret[0].(chan<- *botreq.BlockRequest)
	return ret0
}

// BlockRequestCh indicates an expected call of BlockRequestCh.
func (mr *MockBotClientMockRecorder) BlockRequestCh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockRequestCh", reflect.TypeOf((*MockBotClient)(nil).BlockRequestCh))
}

// Close mocks base method.
func (m *MockBotClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockBotClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockBotClient)(nil).Close))
}

// Closed mocks base method.
func (m *MockBotClient) Closed() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Closed")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Closed indicates an expected call of Closed.
func (mr *MockBotClientMockRecorder) Closed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Closed", reflect.TypeOf((*MockBotClient)(nil).Closed))
}

// CombinationRequestCh mocks base method.
func (m *MockBotClient) CombinationRequestCh() chan<- *botreq.CombinationRequest {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CombinationRequestCh")
	ret0, _ := ret[0].(chan<- *botreq.CombinationRequest)
	return ret0
}

// CombinationRequestCh indicates an expected call of CombinationRequestCh.
func (mr *MockBotClientMockRecorder) CombinationRequestCh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CombinationRequestCh", reflect.TypeOf((*MockBotClient)(nil).CombinationRequestCh))
}

// CombinerBotSubscriptions mocks base method.
func (m *MockBotClient) CombinerBotSubscriptions() []domain.CombinerBotSubscription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CombinerBotSubscriptions")
	ret0, _ := ret[0].([]domain.CombinerBotSubscription)
	return ret0
}

// CombinerBotSubscriptions indicates an expected call of CombinerBotSubscriptions.
func (mr *MockBotClientMockRecorder) CombinerBotSubscriptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CombinerBotSubscriptions", reflect.TypeOf((*MockBotClient)(nil).CombinerBotSubscriptions))
}

// Config mocks base method.
func (m *MockBotClient) Config() config.AgentConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(config.AgentConfig)
	return ret0
}

// Config indicates an expected call of Config.
func (mr *MockBotClientMockRecorder) Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockBotClient)(nil).Config))
}

// Initialize mocks base method.
func (m *MockBotClient) Initialize() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Initialize")
}

// Initialize indicates an expected call of Initialize.
func (mr *MockBotClientMockRecorder) Initialize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialize", reflect.TypeOf((*MockBotClient)(nil).Initialize))
}

// Initialized mocks base method.
func (m *MockBotClient) Initialized() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialized")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Initialized indicates an expected call of Initialized.
func (mr *MockBotClientMockRecorder) Initialized() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialized", reflect.TypeOf((*MockBotClient)(nil).Initialized))
}

// IsClosed mocks base method.
func (m *MockBotClient) IsClosed() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsClosed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsClosed indicates an expected call of IsClosed.
func (mr *MockBotClientMockRecorder) IsClosed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsClosed", reflect.TypeOf((*MockBotClient)(nil).IsClosed))
}

// IsInitialized mocks base method.
func (m *MockBotClient) IsInitialized() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInitialized")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsInitialized indicates an expected call of IsInitialized.
func (mr *MockBotClientMockRecorder) IsInitialized() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInitialized", reflect.TypeOf((*MockBotClient)(nil).IsInitialized))
}

// LogStatus mocks base method.
func (m *MockBotClient) LogStatus() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogStatus")
}

// LogStatus indicates an expected call of LogStatus.
func (mr *MockBotClientMockRecorder) LogStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogStatus", reflect.TypeOf((*MockBotClient)(nil).LogStatus))
}

// SetConfig mocks base method.
func (m *MockBotClient) SetConfig(arg0 config.AgentConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetConfig", arg0)
}

// SetConfig indicates an expected call of SetConfig.
func (mr *MockBotClientMockRecorder) SetConfig(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConfig", reflect.TypeOf((*MockBotClient)(nil).SetConfig), arg0)
}

// ShouldProcessAlert mocks base method.
func (m *MockBotClient) ShouldProcessAlert(event *protocol.AlertEvent) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldProcessAlert", event)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ShouldProcessAlert indicates an expected call of ShouldProcessAlert.
func (mr *MockBotClientMockRecorder) ShouldProcessAlert(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldProcessAlert", reflect.TypeOf((*MockBotClient)(nil).ShouldProcessAlert), event)
}

// ShouldProcessBlock mocks base method.
func (m *MockBotClient) ShouldProcessBlock(blockNumberHex string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldProcessBlock", blockNumberHex)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ShouldProcessBlock indicates an expected call of ShouldProcessBlock.
func (mr *MockBotClientMockRecorder) ShouldProcessBlock(blockNumberHex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldProcessBlock", reflect.TypeOf((*MockBotClient)(nil).ShouldProcessBlock), blockNumberHex)
}

// StartProcessing mocks base method.
func (m *MockBotClient) StartProcessing() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartProcessing")
}

// StartProcessing indicates an expected call of StartProcessing.
func (mr *MockBotClientMockRecorder) StartProcessing() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartProcessing", reflect.TypeOf((*MockBotClient)(nil).StartProcessing))
}

// TxBufferIsFull mocks base method.
func (m *MockBotClient) TxBufferIsFull() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxBufferIsFull")
	ret0, _ := ret[0].(bool)
	return ret0
}

// TxBufferIsFull indicates an expected call of TxBufferIsFull.
func (mr *MockBotClientMockRecorder) TxBufferIsFull() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxBufferIsFull", reflect.TypeOf((*MockBotClient)(nil).TxBufferIsFull))
}

// TxRequestCh mocks base method.
func (m *MockBotClient) TxRequestCh() chan<- *botreq.TxRequest {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxRequestCh")
	ret0, _ := ret[0].(chan<- *botreq.TxRequest)
	return ret0
}

// TxRequestCh indicates an expected call of TxRequestCh.
func (mr *MockBotClientMockRecorder) TxRequestCh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxRequestCh", reflect.TypeOf((*MockBotClient)(nil).TxRequestCh))
}
