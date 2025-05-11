package metadata

import (
	"encoding/json"
	"errors"
	podattr "github.com/Dynatrace/dynatrace-bootstrapper/cmd/configure/attributes/pod"
	"github.com/Dynatrace/dynatrace-operator/pkg/api/v1beta4/dynakube"
	dtwebhook "github.com/Dynatrace/dynatrace-operator/pkg/webhook"
	metacommon "github.com/Dynatrace/dynatrace-operator/pkg/webhook/mutation/pod/common/metadata"
	"maps"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Mutate(metaClient client.Client, request *dtwebhook.MutationRequest, attributes *podattr.Attributes) error {
	log.Info("adding metadata-enrichment to pod", "name", request.PodName())

	workloadInfo, err := metacommon.RetrieveWorkload(metaClient, request)
	if err != nil {
		return err
	}

	attributes.WorkloadInfo = podattr.WorkloadInfo{
		WorkloadKind: workloadInfo.Kind,
		WorkloadName: workloadInfo.Name,
	}

	addMetadataToInitArgs(request, attributes)

	metacommon.SetInjectedAnnotation(request.Pod)
	metacommon.SetWorkloadAnnotations(request.Pod, workloadInfo)

	return nil
}

func addMetadataToInitArgs(request *dtwebhook.MutationRequest, attributes *podattr.Attributes) {
	copiedMetadataAnnotations := metacommon.CopyMetadataFromNamespace(request.Pod, request.Namespace, request.DynaKube)
	if value, ok := request.Pod.Annotations[dynakube.MetadataAnnotation]; ok {
		metadataAnnotations := make(map[string]string)
		err := json.Unmarshal([]byte(value), &metadataAnnotations)
		if err != nil {
			log.Error(err, "yuval failed to marshal annotations to map", "annotations", metadataAnnotations)
		}
		if len(metadataAnnotations) == 0 {
			log.Error(errors.New("yuval metadataAnnotations is empty failed to copy map"), "%+v", request.Pod.Annotations)
			return
		}
		if attributes.UserDefined == nil {
			attributes.UserDefined = make(map[string]string)
		}
		maps.Copy(attributes.UserDefined, metadataAnnotations)
		return
	}
	if copiedMetadataAnnotations == nil {
		log.Error(nil, "yuval copiedMetadataAnnotations is nil failed to copy map")
		return
	}
	if attributes.UserDefined == nil {
		attributes.UserDefined = make(map[string]string)
		log.Error(nil, "yuval attributes.UserDefined is nil so initialized the map then to copy into it")
	}
	maps.Copy(attributes.UserDefined, copiedMetadataAnnotations)
}
