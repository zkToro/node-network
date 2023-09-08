package registry

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"zktoro/zktoro-core-go/domain/registry"
	"zktoro/zktoro-core-go/domain/registry/regmsg"
)

type testHandlerImpl1 struct {
	val string
}

func (impl *testHandlerImpl1) HandleMessage(ctx context.Context, logger *logrus.Entry, msg *registry.AgentSaveMessage) error {
	impl.val = msg.AgentID
	return nil
}

type testHandlerImpl2 struct {
	val string
}

func (impl *testHandlerImpl2) HandleMessage(ctx context.Context, logger *logrus.Entry, msg *registry.DispatchMessage) error {
	impl.val = msg.AgentID
	return nil
}

func TestHandlerRegistry(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()

	testID1 := "0001"
	testID2 := "0002"

	agentSave1 := &testHandlerImpl1{}
	agentSave2 := &testHandlerImpl1{}
	dispatch1 := &testHandlerImpl2{}

	handlerReg := NewHandlerRegistry(Handlers{
		SaveAgentHandlers: regmsg.Handlers(agentSave1.HandleMessage, agentSave2.HandleMessage),
		DispatchHandlers:  regmsg.Handlers(dispatch1.HandleMessage),
	})

	logger := logrus.NewEntry(logrus.StandardLogger())

	err := handlerReg.Handle(ctx, logger, &registry.AgentSaveMessage{
		AgentMessage: registry.AgentMessage{AgentID: testID1},
	})
	r.NoError(err)

	err = handlerReg.Handle(ctx, logger, &registry.DispatchMessage{
		AgentID: testID2,
	})
	r.NoError(err)

	r.Equal(testID1, agentSave1.val)
	r.Equal(testID1, agentSave2.val)
	r.Equal(testID2, dispatch1.val)
}
