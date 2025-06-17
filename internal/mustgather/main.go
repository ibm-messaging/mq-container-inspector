package mustgather

import (
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func MustGather(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// collect pods must-gathers
	err := gatherPodsToFiles(cfg, flags)
	if err != nil {
		return err
	}

	// collect crd must-gathers
	err = gatherCrdToFiles(cfg, flags)
	if err != nil {
		return err
	}

	// collect the route must-gathers
	err = gatherRoutesToFiles(cfg, flags)
	if err != nil {
		return err
	}

	// collect StatefulSet must-gathers
	err = gatherStatefulSetToFiles(cfg, flags)
	if err != nil {
		return err
	}

	// collect Service must-gathers
	err = gatherServicesToFiles(cfg, flags)
	if err != nil {
		return err
	}

	return nil

}
