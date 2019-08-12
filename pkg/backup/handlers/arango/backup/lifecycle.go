package backup

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/rs/zerolog/log"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var _ operator.LifecyclePreStart = &handler{}

// LifecyclePreStart is executed before operator starts to work, additional checks can be placed here
// Wait for CR to be present
func (h *handler) LifecyclePreStart() error {
	log.Info().Msgf("Starting Lifecycle PreStart for %s", h.Name())

	defer func() {
		log.Info().Msgf("Lifecycle PreStart for %s completed", h.Name())
	}()

	for {
		_, err := h.client.DatabaseV1alpha().ArangoBackups("test").List(meta.ListOptions{})

		if err != nil {
			klog.Warningf("CR for %s not found: %s", v1alpha.ArangoBackupResourceKind, err.Error())

			time.Sleep(250 * time.Millisecond)
			continue
		}

		return nil
	}
}
