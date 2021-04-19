
# Image URL to use all building/pushing image targets
RELEASE ?= ""  # Must be set at runtime
IMG ?= armory/pacrd:${RELEASE}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
DOCS_PROJECT ?= ~/armory/documentation
OS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
PWD=$(shell pwd)
server-name=armory-docker-local

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


integration-test-generate-dockerimage: test
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > integration_test/pacrd.yaml
	# docker build -t armory-io/pacrd-integration-test:$(shell date -u +%Y%m%d%H%M%S) -f Dockerfile.integrationtest .
	# docker build -t armory-docker-local.jfrog.io/armory/pacrd-integration-test:latest -f Dockerfile.integrationtest .
	docker build -t pacrd-integration-test:latest -f Dockerfile.integrationtest .

integration-test-run: integration-test-generate-dockerimage
	docker run --privileged pacrd-integration-test:latest

integration-test-push-image: integration-test-generate-dockerimage
	@docker login ${server-name}.jfrog.io -u $(shell grep "artifactory_user" ~/.gradle/gradle.properties |cut -d'=' -f2) -p $(shell grep "artifactory_password" ~/.gradle/gradle.properties |cut -d'=' -f2)
	docker tag pacrd-integration-test:latest ${server-name}.jfrog.io/armory/pacrd-integration-test:latest
	docker push ${server-name}.jfrog.io/armory/pacrd-integration-test:latest

all: manager

# Run tests
test: generate fmt vet manifests
ifeq (, $(wildcard ./kubebuilder/.*))
	curl -L https://go.kubebuilder.io/dl/2.3.1/${OS}/${ARCH} | tar -xz -C .
	mv ./kubebuilder_2.3.1_${OS}_${ARCH} ./kubebuilder
endif
	export PATH=${PATH}:${PWD}/kubebuilder/bin; \
	export KUBEBUILDER_ASSETS=${PWD}/kubebuilder/bin; \
	kubebuilder version; \
	go test -v -mod=vendor -race -covermode atomic -coverprofile=profile.cov ./...; \
	

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go configfile.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go ./configfile.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate the manifest that we'll publish for our customers
generate-public-manifest: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > pacrd-${RELEASE}.yaml
	git add config/manager
	git commit -m "chore(config): update manager version to ${RELEASE}"

# Publish the publicly consumable manifest to one of our properties
publish-public-manifest: generate-public-manifest
	aws --region=us-east-1 --profile=prod s3 cp pacrd-${RELEASE}.yaml s3://engineering.armory.io/manifests/pacrd-${RELEASE}.yaml

# Publish the publicly consumable manifest to one of our properties, this is used in prod for release so no profile is needed
publish-public-manifest-no-profile: generate-public-manifest
	aws --region=us-east-1 s3 cp pacrd-${RELEASE}.yaml s3://engineering.armory.io/manifests/pacrd-${RELEASE}.yaml



# Cache busting; Useful if you need to supply a quick fix to a manifest
invalidate-manifest-cache:
	aws --region=us-east-1 --profile=prod cloudfront create-invalidation --distribution-id ENJEJHR8VKPZA --paths /\*

# List all existing releases
list-manifest-releases:
	aws --region=us-east-1 --profile=prod s3 ls s3://engineering.armory.io/manifests/pacrd-

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

generate-docs: docs/config.json docs/template/
	cp docs/doc.go.tpl api/v1alpha1/doc.go  # The gen-crd tool expects this file; temporarily create it and delete when done.
	gen-crd-api-reference-docs \
		-template-dir docs/template/ \
		-config docs/config.json \
		-api-dir github.com/armory-io/pacrd/api/v1alpha1/ \
		-out-file $(DOCS_PROJECT)/_spinnaker/pacrd-crd-docs.md
	rm api/v1alpha1/doc.go

install-doc-generator:
	go get github.com/ahmetb/gen-crd-api-reference-docs

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

release: generate-public-manifest publish-public-manifest

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
