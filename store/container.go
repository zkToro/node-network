package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"zktoro/zktoro-core-go/release"

	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"

	"zktoro/config"
)

const defaultImageCheckInterval = time.Second * 5

// ZktoroImageStore keeps track of the latest zktoro node image.
type zktoroImageStore interface {
	Latest() <-chan ImageRefs
	EmbeddedImageRefs() ImageRefs
}

// ImageRefs contains the latest image references.
type ImageRefs struct {
	Supervisor  string
	Updater     string
	ReleaseInfo *release.ReleaseInfo
}

type ZktoroImageStore struct {
	updaterPort string
	latestCh    chan ImageRefs
	latestImgs  ImageRefs
}

// NewZktoroImageStore creates a new store.
func NewZktoroImageStore(ctx context.Context, updaterPort string, autoUpdate bool) (*ZktoroImageStore, error) {
	store := &ZktoroImageStore{
		updaterPort: updaterPort,
		latestCh:    make(chan ImageRefs),
	}
	if autoUpdate {
		go store.loop(ctx)
	}
	return store, nil
}

func (store *ZktoroImageStore) loop(ctx context.Context) {
	store.check(ctx)
	ticker := time.NewTicker(defaultImageCheckInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			store.check(ctx)
		}
	}
}

func (store *ZktoroImageStore) EmbeddedImageRefs() ImageRefs {
	return ImageRefs{
		Supervisor:  config.DockerSupervisorImage,
		Updater:     config.DockerUpdaterImage,
		ReleaseInfo: config.GetBuildReleaseInfo(),
	}
}

func (store *ZktoroImageStore) check(ctx context.Context) {
	latestReleaseInfo, err := store.getFromUpdater(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to get the latest release from the updater")
	}

	if latestReleaseInfo == nil {
		return
	}

	serviceImgs := latestReleaseInfo.Manifest.Release.Services
	if serviceImgs.Supervisor != store.latestImgs.Supervisor || serviceImgs.Updater != store.latestImgs.Updater {
		log.WithField("commit", latestReleaseInfo.Manifest.Release.Commit).Info("got newer release from updater")

		store.latestImgs = ImageRefs{
			Supervisor:  serviceImgs.Supervisor,
			Updater:     serviceImgs.Updater,
			ReleaseInfo: latestReleaseInfo,
		}
		store.latestCh <- store.latestImgs
	}
}

func (store *ZktoroImageStore) getFromUpdater(ctx context.Context) (*release.ReleaseInfo, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s", store.updaterPort))
	if err != nil {
		return nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound { // 404 == not ready yet
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected updater response with code %d: %s", resp.StatusCode, string(respBody))
	}
	var releaseInfo release.ReleaseInfo
	return &releaseInfo, json.Unmarshal(respBody, &releaseInfo)
}

// Latest returns a channel that provides the latest image reference.
func (store *ZktoroImageStore) Latest() <-chan ImageRefs {
	return store.latestCh
}
