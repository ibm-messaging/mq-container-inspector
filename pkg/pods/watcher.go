/*
© Copyright IBM Corporation 2025

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pods

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchtools "k8s.io/client-go/tools/watch"
)

func SetupPodWatcher(client kubernetes.Interface, pod *corev1.Pod, namespace, watchFor string, logger *slog.Logger) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fieldSelector := fields.OneTermEqualSelector("metadata.name", pod.ObjectMeta.Name).String()

	watcher := getPodListWatcher(client, ctx, fieldSelector, namespace)

	watchCondition := getPodWatcherCondition(watchFor, logger)

	_, err := watchtools.UntilWithSync(ctx, watcher, &corev1.Pod{}, nil, watchCondition)
	return err

}

func getPodWatcherCondition(watchFor string, logger *slog.Logger) func(watch.Event) (bool, error) {

	switch watchFor {
	case utils.PodCreationWatcher:
		return func(watchEvent watch.Event) (bool, error) {

			pod := watchEvent.Object.(*corev1.Pod)

			// done if reached terminal phase
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				fmt.Printf("PVC-inspector pod %s reached terminal phase: %s\n", pod.ObjectMeta.Name, pod.Status.Phase)
				logger.Info(fmt.Sprintf("PVC pod %s reached terminal phase: %s", pod.ObjectMeta.Name, pod.Status.Phase))
				return true, nil
			}

			// done when the pod is ready
			for _, podCondition := range pod.Status.Conditions {
				if podCondition.Type == corev1.PodReady && podCondition.Status == corev1.ConditionTrue {
					fmt.Printf("PVC-inspector pod %s is ready\n", pod.ObjectMeta.Name)
					logger.Info(fmt.Sprintf("PVC pod %s is ready", pod.ObjectMeta.Name))
					return true, nil
				}
			}

			return false, nil

		}

	case utils.PodDeletionWatcher:
		return func(watchEvent watch.Event) (bool, error) {

			pod, ok := watchEvent.Object.(*corev1.Pod)

			if !ok {
				return false, nil
			}

			if watchEvent.Type == watch.Deleted {
				fmt.Printf("PVC-inspector pod %s deleted\n", pod.ObjectMeta.Name)
				logger.Info(fmt.Sprintf("PVC-inspector pod %s deleted", pod.ObjectMeta.Name))
				return true, nil
			}

			if pod.DeletionTimestamp != nil {
				fmt.Printf("PVC-inspector pod %s is terminating\n", pod.ObjectMeta.Name)
				logger.Info(fmt.Sprintf("PVC-inspector pod %s is terminating", pod.ObjectMeta.Name))
			}

			return false, nil
		}
	}

	return nil

}

func getPodListWatcher(client kubernetes.Interface, ctx context.Context, fieldSelector, namespace string) *cache.ListWatch {

	return &cache.ListWatch{

		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				FieldSelector: fieldSelector,
			})
		},

		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
				FieldSelector: fieldSelector,
			})
		},
	}

}
