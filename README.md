## Getting Started

The following will describe how to install and configure the DigitalOcean block store plugin for Velero and provide a usage example.

* [Prerequisites](#prerequisites)
* [Quickstart](#quickstart)
* [Block store](#block-store)
* [Object store](#object-store)
* [Backup and restore example](#backup-and-restore-example)
* [Build image](#build-image)

### Prerequisites

* [Kubernetes cluster](https://stackpoint.io/clusters/new?provider=do)
* DigitalOcean account and resources
  * [API personal access token](https://www.digitalocean.com/docs/api/create-personal-access-token/)
  * [Spaces access keys](https://www.digitalocean.com/docs/spaces/how-to/administrative-access/)
  * Spaces bucket
  * Spaces bucket region
* [Velero](https://heptio.github.io/velero/master/) v0.11.x prerequisites

### Quickstart

This quickstart will describe the installation and configuration of the DigitalOcean block store plugin for Velero as well as the built-in object store using DigitalOcean Spaces. Please review the [Block store](#block-store) and [Object store](#object-store) sections further down in the README for more details on each component.

1. Complete the Heptio Velero prerequisites mentioned above. This generally involves applying the `00-prereqs.yaml` available from the Velero repository:

    ```
    kubectl apply -f examples/00-prereqs.yaml
    ```

2. Update the `examples/credentials-velero` with your Spaces access and secret keys. The file will look like the following:

    ```
    [default]
    aws_access_key_id=<AWS_ACCESS_KEY_ID>
    aws_secret_access_key=<AWS_SECRET_ACCESS_KEY>
    ```

3. Create a Kubernetes `cloud-credentials` secret containing the `credentials-velero` and DigitalOcean API token.

    ```
    kubectl create secret generic cloud-credentials \
        --namespace velero \
        --from-file cloud=examples/credentials-velero \
        --from-literal digitalocean_token=<DIGITALOCEAN_TOKEN>
    ```

4. Update the `examples/05-velero-backupstoragelocation.yaml` with the DigitalOcean Spaces API URL, bucket, and region and apply the `BackupStorageLocation` configuration. The `BackupStorageLocation` uses the AWS S3-compatible provider to communicate with DigitalOcean Spaces.

    ```
    kubectl apply -f examples/05-velero-backupstoragelocation.yaml
    ```

5. Next apply the `VolumeSnapshotLocation` configuration. No updates are required to the YAML.

    ```
    kubectl apply -f examples/06-velero-volumesnapshotlocation.yaml
    ```

6. Now apply the Velero deployment.

    ```
    kubectl apply -f examples/10-deployment.yaml
    ```

7. Finally add the `velero-blockstore-digitalocean` plugin to Velero.

    ```
    velero plugin add gcr.io/stackpoint-public/velero-blockstore-digitalocean:latest
    ```

### Block store

The block store provider manages snapshots for DigitalOcean persistent volumes.

1. The block store provider requires a personal access token to create and restore snapshots through the DigitalOcean API. This token can be generated through the DigitalOcean Control Panel as describe [here](https://www.digitalocean.com/docs/api/create-personal-access-token/).

2. Once the token is available, create a Secret using the new token.

    ```
    kubectl create secret generic cloud-credentials \
        --namespace velero \
        --from-literal digitalocean_token=<DIGITALOCEAN_TOKEN>
    ```

3. Velero must be aware of the cloud provider to use with persistent volumes. You can create custom snapshot location by running.

    ```
	velero snapshot-location create do-blockstore --provider digitalocean-blockstore
    ```

4. Next the Deployment should be updated with the `cloud-credentials` Secret.

    ```
    kubectl -n velero edit deployment velero
    ```

    A full Deployment YAML example defining the Secret can be found in `examples/20-deployment.yaml`.

5. Finally, add the `velero-blockstore-digitalocean` plugin to Velero.

    ```
    velero plugin add gcr.io/stackpoint-public/velero-blockstore-digitalocean:latest
    ```

### Object store

The object store uses [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) to store the backup files. As Spaces is an S3-compatible object storage solution, the object store will use the Velero built-in `aws` provider.

1. First generate the Spaces access key and secret key in the DigitalOcean Control Panel as described [here](https://www.digitalocean.com/docs/spaces/how-to/administrative-access/).

2. A Spaces bucket must also be created through the DigitalOcean Control Panel before proceeding with Velero configuration. Make note of the bucket name and region as these will be required later.

3. Once the access and secret keys are available, create an S3-compatible `credentials-velero` file with the new keys.

    ```
    [default]
    aws_access_key_id=<DO_ACCESS_KEY_ID>
    aws_secret_access_key=<DO_SECRET_ACCESS_KEY>
    ```

4. The `credentials-velero` file must then be added to the `cloud-credentinals` Secret:

    ```
    kubectl create secret generic cloud-credentials \
        --namespace velero \
        --from-file cloud=./credentials-velero
    ```

5. Now create DigitalOcean backup location:

    ```
	velero backup-location create do-spaces --provider aws
    ```

	and then edit it setting region, s3Url and bucket to correct values:
	
	```
	kubectl -n velero edit backupstoragelocations do-spaces
	```

6. Finally, the Deployment can be updated with the `cloud-credentials` Secret.

    ```
    kubectl -n velero edit deployment velero
    ```

    A full Deployment YAML example can be found in `examples/10-deployment.yaml`.


### Backup and restore example

1. Apply the Nginx `examples/nginx-pv.yml` config that uses persistent storage for the log path.

    ```
    kubectl apply -f examples/nginx-pv.yml
    ```

2. Once Nginx deployment is running and available, create a backup using Velero.

    ```
    velero backup create nginx-backup --selector app=nginx
    velero backup describe nginx-backup
    ```

3. The config files should appear in the Spaces bucket and a snapshot taken of the persistent volume. Now you can simulate a disaster by deleting the `nginx-example` namespace.

    ```
    kubectl delete namespace nginx-example
    ```

4. The `nginx-data` backup can now be restored.

    ```
    velero restore create --from-backup nginx-backup
    ```

### Build image

```
make clean
make container IMAGE=gcr.io/stackpoint-public/velero-blockstore-digitalocean:devel
```
