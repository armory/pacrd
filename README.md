[![experiment](https://img.shields.io/badge/status-experiment-yellowgreen?style=flat-square)](https://kb.armory.io/releases/early-release-beta-GA/)

# pacrd

PaCRD (a combination of "Pipelines as Code" and "Custom Resource Definition") is
a [Kubernetes controller] that manages the lifecycle of Spinnaker applications
and pipelines as objects within your cluster. PaCRD extends Kubernetes
functionality to support Spinnaker Application and Pipeline objects that can be
observed for changes through a mature lifecycle management API.

## Development

### Pre-Requisites

- A Kubernetes cluster
- Go >= 1.13  
- Kustomize

### Installing and updating CRDs During Development

CRD definitions are generated and stored in the `config/crd/bases/` directory.
In order for your controller to pick up changes in your `api/` models you will
need to re-generatre them and apply them to the target cluster. You can most
easily do this with the following command:

```
make manifests &&\
  kubectl apply -f /config/crd/bases/
```

You should see something like the following output:

```
/Users/my-user/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
customresourcedefinition.apiextensions.k8s.io/applications.pacrd.armory.spinnaker.io configured
customresourcedefinition.apiextensions.k8s.io/pipelines.pacrd.armory.spinnaker.io configured
```

### Local Development

When making and testing changes to the controller locally it's helpful to be
able to run the controller and connect to a running Kubernetes cluster. In
order to run the controller locally you will need to:

- Create a configuration file for PaCRD:

```yaml
# file: /path/to/pacrd/pacrd.yaml
spinnakerServices:
  front50: http://localhost:8080
  orca: http://localhost:8083
```

- Port-forward Front50 and Orca to your local machine:

```sh
kubectl port-forward service/spin-front50 8080
```

```sh
kubectl port-forward service/spin-orca 8083
```

- (Optional) Port-forward Deck and Gate to your local machine:

```sh
kubectl port-forward service/spin-gate 8084
```

```sh
kubectl port-forward service/spin-deck 9000
```

- Start the controller:

```sh
make run
```

If you've configured your environment successfully you should see the
following output:

```
/Users/my-user/bin/controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."
go fmt ./...
go vet ./...
/Users/my-user/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
go run ./main.go ./configfile.go
2020-03-20T08:20:12.709-0700    INFO    setup   Initializing PaCRD configuration
2020-03-20T08:20:13.910-0700    INFO    controller-runtime.metrics      metrics server is starting to listen    {"addr": ":8080"}
2020-03-20T08:20:13.911-0700    INFO    setup   starting manager
2020-03-20T08:20:13.911-0700    INFO    controller-runtime.manager      starting metrics server {"path": "/metrics"}
2020-03-20T08:20:14.016-0700    INFO    controller-runtime.controller   Starting EventSource    {"controller": "application", "source": "kind source: /, Kind="}
2020-03-20T08:20:14.016-0700    INFO    controller-runtime.controller   Starting EventSource    {"controller": "pipeline", "source": "kind source: /, Kind="}
2020-03-20T08:20:14.120-0700    INFO    controller-runtime.controller   Starting Controller     {"controller": "pipeline"}
2020-03-20T08:20:14.120-0700    INFO    controller-runtime.controller   Starting Controller     {"controller": "application"}
2020-03-20T08:20:14.221-0700    INFO    controller-runtime.controller   Starting workers        {"controller": "application", "worker count": 1}
2020-03-20T08:20:14.221-0700    INFO    controller-runtime.controller   Starting workers        {"controller": "pipeline", "worker count": 1}
2020-03-20T08:20:14.221-0700    INFO    controllers.Application reconciling application {"application": "my-user/api-server"}
```

### Cutting a Manual Release

At this time releases for PaCRD are done manually, though the desired
end-state is to have automated builds/releases. The overall processes looks
like this:

1. Build and publish a PaCRD container from your local machine
1. Build and publish the PaCRD manifest from your local machine
1. Update documentation if needed

#### Publish PaCRD Container

You can build and publish the PaCRD container assuming you've logged into
DockerHub and your user has access to the `armory` account. You can build
and publish the container like so:

```sh
make docker-build docker-publish
```

This will publish a PaCRD container with the current git commit sha as the label.

#### Publish PaCRD Manifest

If you need to find out what the current releases are you can run the following
`make` target:

```
make list-manifest-releases
```

You should see output like the following:

```
aws --region=us-east-1 --profile=prod s3 ls s3://engineering.armory.io/manifests/pacrd-
2020-03-10 12:53:48      32000 pacrd-0.1.0.yaml
2020-03-11 15:13:32      32063 pacrd-0.1.1.yaml
```

Next, release a new version of the manifest with the following `make` target,
where `RELEASE` is set to the desired version:

```
make publish-public-manifest RELEASE=0.2.0
```

You should see output like the following:

```
/Users/my-user/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
cd config/manager && kustomize edit set image controller=armory/pacrd:61bb144-dirty
kustomize build config/default > pacrd-0.2.0.yaml
aws --region=us-east-1 --profile=prod s3 cp pacrd-0.2.0.yaml s3://engineering.armory.io/manifests/pacrd-0.2.0.yaml
upload: ./pacrd-0.2.0.yaml to s3://engineering.armory.io/manifests/pacrd-0.2.0.yaml
```

### Generating API Documentation

This project has the ability to generate human-readable documentation for
inclusion in Armory's doc site.

We make the following assumptions:

- You have the `gen-crd-api-reference-docs` tool installed
  - If you don't, run `make install-doc-generator`
- `$GOBIN` is defined and in your `$PATH`
- Your checkout of the Armory docs project is at `~/armory/documentation`
  - If it is not, you can set the `DOCS_PROJECT` variable when running the generation command

To generate documentation, run the following `make` target:

```
make generate-docs  # optionally supply DOCS_PROJECT=your/docs/checkout/path
```

Then create a pull-request to the docs project with your updates!
