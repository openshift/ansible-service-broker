# Portable Service Assets

## Introduction
Service classes in the Service Catalog require a number of UI graphical assets (i.e. service icon, company logo, etc).  For Ansible Playbook Bundles (APBs) that are sourced from the RHCC, we do not want to reference these externally on a remote (non-RH managed endpoint).  Additionally, we want to be able to handle off-network / disconnected clusters thus making this a truly portable solution.

## Problem Description
Allow APBs to specify static file assets which will be accessible to any computer using OpenShift.
* How do we build and store these assets for each APB?
* How do we download these assets before downloading the entire APB since they will be needed before deploying?
* How do we serve these assets?

## Building and Storing Assets
Assets will be stored under the APB folder in a folder named `assets`. As part of the `apb build` process, a container image will be created copying all files from the `assets` folder into the container.

```Dockerfile
# Assets Dockerfile
FROM scratch
COPY assets /assets
CMD ["noop"]
```

This docker container will be built with the same image name, but the tag will be prefixed with `_assets_`.  APB tooling will perform `docker build -t hello-world-apb:_assets_latest .`  and keep track of the image ID.

After building the assets image, the tooling will also add a label to the APB image with the assets image ID.  This will ensure that if the APB is updated,  the matching assets will be associated and we can detect changes in the assets even if the APB spec has not changed.

```bash
docker build . -t docker.io/cfchase/hello-world-apb:latest --label "com.redhat.apb.assetsImageId=<ImageId>"
```

## Specifying Assets
In a given `apb.yml` the user can specify `{{ assets }}` inside any metadata string values. Currently, only metadata is planned for support, but we could parse more if the use case is made.

```yaml
#apb.yml
name: hello-world-apb
...
metadata:
  imageUrl: "{{ assets }}/my-pic.svg"
  documentationUrl: "{{ assets }}/docs.html"
```

## Downloading Assets
After every bootstrap of apb specs, ASB will begin downloading assets in an asynchronous work job.  For each spec, Ansible Service Broker (ASB) will check to see if there is an associated assets image ID.  If there is none, no assets are downloaded.  If there is an assets image ID and the assets are missing or the ID has changed from the last download, it will download them from the built image assets (e.g. `hello-world-apb:_assets_latest`) using commands mimicking the following:
```bash
mkdir -p <www-docroot>/assets/apbs/dh-ansibleplaybookbundle-hello-world-apb-latest/
docker create docker.io/ansibleplaybookbundle/hello-world-apb:_assets_latest
docker cp $DOWNLOADED_IMAGE_ID:/assets/. <www-docroot>/assets/apbs/dh-ansibleplaybookbundle-hello-world-apb-latest
```
To avoid using the docker daemon, we will attempt to use [buildah](https://github.com/projectatomic/buildah) to perform those commands.

## Serving Assets
ASB will serve files from a document root.  This folder would be specified by the broker config and defaulted to `/var/www`. This folder should be available without authentication.

The catalog request will be modified to supply the path to assets. The `{{ assets }}` string in the spec can be replaced with the URL for the assets for that specific APB as part of the catalog request, `<asb-route>/public/assets/apbs/<spec-fq-name>`.  To do this we would need to look up the the route to ASB and add the information specific to each spec.  We can look for a route passed in using the broker config.  If missing, we can try to query for it on some best guesses.
```bash
# route name provided:
oc get routes <brokerConfig.RouteName> -n <brokerConfig.Namespace> -o custom-columns=host:.spec.host --no-headers
# or if no route name, assume service is named asb:
oc get routes -n <brokerConfig.Namespace> -l service=asb -o custom-columns=host:.spec.host --no-headers
```

The resulting request would look like:
```json
[
     {
          "name": "dh-ansibleplaybookbundle-hello-world-apb-latest",
          "metadata": {
             "imageUrl": "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/public/assets/apbs/dh-ansibleplaybookbundle-hello-world-apb-latest/my-pic.svg",
             "documentationUrl": "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/public/assets/apbs/dh-ansibleplaybookbundle-hello-world-apb-latest/docs.html"
         }
     }
 ]
```
This will then be used by the service catalog as any normal URI and the asset will be served from ASB.

## Work Items
* APB tooling to build assets
* Serving files
* Downloading assets
* Modify APB processing for requests
* Update documentation for building APBs

## Future Considerations
For running multiple instances of the broker, we'll have to eventually use a persistent volume and consider multiple pods downloading assets.  We could also explore using a separate pod with httpd in conjunction with the persistent volume instead of serving assets from the broker.
