# © Copyright IBM Corporation 2025
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include config.env

####################################
# Variables used
####################################
PLATFORM_TAG=${OS}-${ARCH}
# Contruct the name of binary based on os & architecture
BINARY_FULL_NAME=${MQ_INSPECTOR_VERSION}_${BINARY}_${PLATFORM_TAG}
# Binary name without version
BINARY_FULL_NAME_LATEST=${BINARY}_${PLATFORM_TAG}
# Contruct a folder name based on the commit id & timestamp
SANITIZED_TIMESTAMP=$(subst :,_,${COMMIT_TIMESTAMP})
# Cut the commit id to 7 characters 
SHORT_ID=$(shell echo $(COMMITID) | cut -c1-7)
COMMIT_FOLDER_NAME=${SANITIZED_TIMESTAMP}_${SHORT_ID}
# Names for the tars to be generated
# Tar for main folder
TAR_NAME_MAIN=${BINARY_FULL_NAME}.tar.gz
# .tar name without version
TAR_NAME_LATEST=${BINARY_FULL_NAME_LATEST}.tar.gz
 
###############################################################################
# Build targets
###############################################################################

.PHONY: build-mainline
build-mainline: build-multi-architecture tar-binary push-artifact-mainline

.PHONY: tar-binary
tar-binary: tar-binary-main tar-binary-latest

# Target to push for a mainline build
.PHONY: push-artifact-mainline
push-artifact-mainline: push-artifact-main-builds push-artifact-latest

# Build for any combination of OS and ARCH
build-multi-architecture:
	GOOS=${OS} GOARCH=${ARCH} go build -o ${BINARY_FULL_NAME}

tar-binary-main:
	tar -czf ${TAR_NAME_MAIN} ${BINARY_FULL_NAME}

tar-binary-latest:
	mv ${BINARY_FULL_NAME} ${BINARY_FULL_NAME_LATEST}
	tar -czf ${TAR_NAME_LATEST} ${BINARY_FULL_NAME_LATEST}

# Push tar to folder by name latest
push-artifact-latest:
	curl -u "${INSPECTOR_REGISTRY_USER}:${INSPECTOR_REGISTRY_CREDENTIAL}" -T ${TAR_NAME_LATEST} "${INSPECTOR_REGISTRY_HOSTNAME}/${INSPECTOR_REGISTRY_NAMESPACE}/${REGISTRY_FOLDER_LATEST}/"

# Push tar to the dynamically created COMMIT_FOLDER under main-builds folder
push-artifact-main-builds:
	curl -u "${INSPECTOR_REGISTRY_USER}:${INSPECTOR_REGISTRY_CREDENTIAL}" -T ${TAR_NAME_MAIN} "${INSPECTOR_REGISTRY_HOSTNAME}/${INSPECTOR_REGISTRY_NAMESPACE}/${REGISTRY_FOLDER_MAIN}/${COMMIT_FOLDER_NAME}/"
