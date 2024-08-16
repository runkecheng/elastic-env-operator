package common

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const CronHPAIndexByRef = "byRefWorkload"

func BuildRefKey(obj client.Object) string {
	return fmt.Sprintf("%s/%s.%s.%s.%s", obj.GetObjectKind().GroupVersionKind().Group,
		obj.GetObjectKind().GroupVersionKind().Version,
		obj.GetObjectKind().GroupVersionKind().Kind,
		obj.GetNamespace(),
		obj.GetName())
}
