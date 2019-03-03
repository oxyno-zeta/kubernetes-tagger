# Data Structure

This page will show the available data structure for queries or conditions.

## Root

| Key                   | Description                                                                                                                                |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| type                  | Resource type (for example: "volume")                                                                                                      |
| platform              | Resource platform (for example: "aws")                                                                                                     |
| persistentvolume      | [PersistentVolumeStructure](#persistentvolumestructure) (Only if the resource is a persistent volume)                                      |
| persistentvolumeclaim | [PersistentVolumeClaimStructure](#persistentvolumeclaimstructure) (Only when a persistent volume claim is linked to the persistent volume) |

## PersistentVolumeStructure

| Key              | Description                                                                                    |
| ---------------- | ---------------------------------------------------------------------------------------------- |
| labels           | This is the `map[string]string` got from `labels` in the Kubernetes PersistentVolume Kind      |
| annotations      | This is the `map[string]string` got from `annotations` in the Kubernetes PersistentVolume Kind |
| name             | The PersistentVolume name                                                                      |
| phase            | The PersistentVolume status phase                                                              |
| reclaimpolicy    | The PersistentVolume Spec Reclaim Policy                                                       |
| storageclassname | The PersistentVolume storage class name                                                        |

## PersistentVolumeClaimStructure

| Key         | Description                                                                                         |
| ----------- | --------------------------------------------------------------------------------------------------- |
| labels      | This is the `map[string]string` got from `labels` in the Kubernetes PersistentVolumeClaim Kind      |
| annotations | This is the `map[string]string` got from `annotations` in the Kubernetes PersistentVolumeClaim Kind |
| namespace   | The PersistentVolumeClaim namespace                                                                 |
| name        | The PersistentVolumeClaim name                                                                      |
| phase       | The PersistentVolumeClaim Status phase                                                              |
