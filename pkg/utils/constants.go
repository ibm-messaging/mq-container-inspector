/*
© Copyright IBM Corporation 2025, 2026

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
package utils

const TimestampFormat = "20060102_150405"

const MustGather = "mustgather"

const PVCInspector = "pvctool"

const MQAgents = "mq-agent"

const Version = "version"

const ApiVersionV1 = "v1"

const ApiVersionAppsV1 = "apps/v1"

const ApiVersionRouteV1 = "route.openshift.io/v1"

const ApiVersionNetworkingV1 = "networking.k8s.io/v1"

const KindPod = "Pod"

const KindList = "List"

const KindStatefulSet = "StatefulSet"

const KindQueueManager = "QueueManager"

const KindControllerRevision = "ControllerRevision"

const KindRoute = "Route"

const KindIngress = "Ingress"

const KindService = "Service"

const KindServiceAccount = "ServiceAccount"

const KindPVC = "PersistentVolumeClaim"

const KindDeployment = "Deployment"

const KindReplicaSet = "ReplicaSet"

const KindDaemonSet = "DaemonSet"

const KindCSV = "ClusterServiceVersion"

const KindCSVList = "ClusterServiceVersionList"

const KindIntegrationKeycloakClientList = "KindIntegrationKeycloakClient"

const KindIntegrationKeycloakClient = "IntegrationKeycloakClient"

const KindCRD = "CustomResourceDefinition"

const KindConfigMap = "ConfigMap"

const KindNetworkPolicy = "NetworkPolicy"

const QmgrGroup = "mq.ibm.com"

const QmgrVersion = "v1beta1"

const QmgrResource = "queuemanagers"

const QmgrContainer = "qmgr"

const PVCInspectorContainer = "pvc-inspector"

const IntegrationKeycloakClientGroup = "keycloak.integration.ibm.com"

const IntegrationKeycloakClientVersion = "v1beta1"

const IntegrationKeycloakClientResource = "integrationkeycloakclients"

const RouteAPIGroupName = "route.openshift.io"

const OperatorGroup = "operators.coreos.com"

const OperatorVersion = "v1alpha1"

const CSVResource = "clusterserviceversions"

const GlobalOperatorNamespace = "openshift-operators"

const ResourcePod = "pods"

const ExecSubcommand = "exec"

const CommonServicesOperatorPrefix = "ibm-common-service-operator"

const CP4iOperatorPrefix = "ibm-integration-platform-navigator"

const WebConsoleLogPath = "var/mqm/web/installations/Installation1/servers/mqweb/logs/console.log"

const WebConsoleMessageLogPath = "var/mqm/web/installations/Installation1/servers/mqweb/logs/messages.log"

const MustGatherLogFileName = "must-gather-logs.log"

const PVCInspectorLogFileName = "pvc-inspector-logs.log"

const MQAgentMustGatherLogFileName = "mq-agent-must-gather-logs.log"

const CRDGroup = "apiextensions.k8s.io"

const CRDResource = "customresourcedefinitions"

const NativeHAEnvName = "MQ_NATIVE_HA"

const MultiInstanceEnvName = "MQ_MULTI_INSTANCE"

const QueueManagerEnvName = "MQ_QMGR_NAME"

const NativeHA = "NativeHA"

const MultiInstance = "MultiInstance"

const SingleInstance = "SingleInstance"

const NodeAffinityHostNameKey = "kubernetes.io/hostname"

const PodCreationWatcher = "PodCreation"

const PodDeletionWatcher = "PodDeletion"

const PodRunningStatus = "Running"

const CustomISAConfigMap = "mq-container-inspector-runmqras-config"

const CustomISAFileName = "custom-isa.xml"

const MQInspectorManagedLabel = "app.kubernetes.io/managed-by:mq-container-inspector"

const QueueManagerStatusPending = "Pending"

const MQAgentLabels = "app.kubernetes.io/managed-by=Helm,app.kubernetes.io/component=integration"

const MQAgentManagedByLabel = "app.kubernetes.io/managed-by=Helm"

const HelmReleaseNameAnnotation = "meta.helm.sh/release-name"

const HelmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
