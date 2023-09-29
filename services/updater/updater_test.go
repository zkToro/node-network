package updater

import (
	"context"
	"testing"

	"zktoro/store"
	mock_store "zktoro/store/mocks"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
)

const (
	testUpdateCheckIntervalSeconds = 1
	testUpdateDelaySeconds         = 15
)

func TestUpdaterService_UpdateLatestRelease(t *testing.T) {
	r := require.New(t)

	svs := mock_store.NewMockScannerReleaseStore(gomock.NewController(t))
	updater := NewUpdaterService(
		context.Background(), svs, "8080", testUpdateDelaySeconds, testUpdateCheckIntervalSeconds,
	)

	svs.EXPECT().GetRelease(gomock.Any()).Return(&store.ScannerRelease{
		Reference: "reference",
	}, nil).Times(2)

	err := updater.updateLatestReleaseWithDelay(0)
	r.NoError(err)
}

func TestUpdaterService_UpdateLatestRelease_SingleEachTime(t *testing.T) {
	r := require.New(t)

	svs := mock_store.NewMockScannerReleaseStore(gomock.NewController(t))
	updater := NewUpdaterService(
		context.Background(), svs, "8080", testUpdateDelaySeconds, testUpdateCheckIntervalSeconds,
	)

	svs.EXPECT().GetRelease(gomock.Any()).Return(&store.ScannerRelease{
		Reference: "reference1",
	}, nil).Times(2)

	svs.EXPECT().GetRelease(gomock.Any()).Return(&store.ScannerRelease{
		Reference: "reference2",
	}, nil).Times(2)

	r.NoError(updater.updateLatestReleaseWithDelay(0))
	r.Equal("reference1", updater.latestReference)

	r.NoError(updater.updateLatestReleaseWithDelay(0))
	r.Equal("reference2", updater.latestReference)
}

func TestUpdaterService_UpdateLatestRelease_TwoInARow(t *testing.T) {
	r := require.New(t)

	svs := mock_store.NewMockScannerReleaseStore(gomock.NewController(t))
	updater := NewUpdaterService(
		context.Background(), svs, "8080", testUpdateDelaySeconds, testUpdateCheckIntervalSeconds,
	)

	finalRef := "reference2"

	svs.EXPECT().GetRelease(gomock.Any()).Return(&store.ScannerRelease{
		Reference: "reference1",
	}, nil).Times(1)

	svs.EXPECT().GetRelease(gomock.Any()).Return(&store.ScannerRelease{
		Reference: "reference2",
	}, nil).Times(1)

	r.NoError(updater.updateLatestReleaseWithDelay(updater.updateDelay))

	// should update to the latest one
	r.Equal(finalRef, updater.latestReference)
}
