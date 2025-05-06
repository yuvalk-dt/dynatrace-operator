package metadata

import (
	"encoding/json"
	"github.com/vladimirvivien/gexe/str"
	"strings"

	"github.com/Dynatrace/dynatrace-operator/pkg/api/v1beta4/dynakube"
	corev1 "k8s.io/api/core/v1"
)

func CopyMetadataFromNamespace(pod *corev1.Pod, namespace corev1.Namespace, dk dynakube.DynaKube) {
	copyAccordingToCustomRules(pod, namespace, dk)
	copyAccordingToPrefix(pod, namespace)
}

func copyAccordingToPrefix(pod *corev1.Pod, namespace corev1.Namespace) {
	for key, value := range namespace.Annotations {
		if strings.HasPrefix(key, dynakube.MetadataPrefix) {
			setPodAnnotationIfNotExists(pod, key, value)
		}
	}
}

func copyAccordingToCustomRules(pod *corev1.Pod, namespace corev1.Namespace, dk dynakube.DynaKube) {
	dataForJsonCopy := map[string]string{}
	for _, rule := range dk.Status.MetadataEnrichment.Rules {
		var key string
		var valueFromNamespace string
		var exists bool

		switch rule.Type {
		case dynakube.EnrichmentLabelRule:
			valueFromNamespace, exists = namespace.Labels[rule.Source]
		case dynakube.EnrichmentAnnotationRule:
			valueFromNamespace, exists = namespace.Annotations[rule.Source]
		}

		key = rule.Target
		if str.IsEmpty(key) {
			key = rule.Source
		}

		if exists {
			enrichmentKey := getMetadataEnrichmentKey(dynakube.EnrichmentNamespaceKey, string(rule.Type), key)
			dataForJsonCopy[enrichmentKey] = valueFromNamespace
		}
	}

	jsonAnnotations, err := json.Marshal(dataForJsonCopy)
	if err != nil {
		log.Info("failed to marshal annotations to map ", err)
		return
	}

	setPodAnnotationIfNotExists(pod, dynakube.MetadataAnnotation, string(jsonAnnotations))
}

func setPodAnnotationIfNotExists(pod *corev1.Pod, key, value string) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	if _, ok := pod.Annotations[key]; !ok {
		pod.Annotations[key] = value
	}
}

func getMetadataEnrichmentKey(enrichmentKey, metadataType, key string) string {
	return enrichmentKey + strings.ToLower(metadataType) + "." + key
}
