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
BINARY=mq-container-inspector
PLATFORM_TAG=${OS}-${ARCH}
# Contruct the name of binary based on os & architecture
BINARY_FULL_NAME=${MQ_INSPECTOR_VERSION}_${BINARY}_${PLATFORM_TAG}
# Names for the tars to be generated
# Tar for main folder
TAR_NAME=${BINARY_FULL_NAME}.tar.gz
 
###############################################################################
# Build targets
###############################################################################

# Build for any combination of OS and ARCH
build-multi-architecture:
	GOOS=${OS} GOARCH=${ARCH} go build -o ${BINARY_FULL_NAME}

.PHONY: tar-binary
tar-binary:
	tar -czf ${TAR_NAME} ${BINARY_FULL_NAME}
