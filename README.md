# azbloblease
Tool to be mainly used from a script to obtain an Azure Storage Blob lease.

For usage example on bash, please see [sample script](./sample/samplelease.sh).

## More examples

### Custom Cloud

``` bash
# Create blob to be lease
./azbloblease createleaseblob -accountname "<storage account name>" -container "azbloblease" -blobname "myblob" -resourcegroupname "<resource group name>" -subscriptionid "<subscription id>" -environment CUSTOMCLOUD

# Lease blob
LEASEID=$(./azbloblease acquire -accountname "<storage account name>" -container "azbloblease" -blobname "myblob" -resourcegroupname "pmarques-rg" -subscriptionid "<subscription id>" -environment CUSTOMCLOUD -leaseduration 60 -custom-cloudconfig-file /tmp/usgov.json | jq -r ".leaseId")

# Maintain lease
./azbloblease renew -accountname "<storage account name>" -container "azbloblease" -blobname "myblob" -resourcegroupname "pmarques-rg" -subscriptionid "<subscription id>" -environment CUSTOMCLOUD -iterations 10 -leaseid $LEASEID -custom-cloudconfig-file /tmp/usgov.json
```

### Custom Cloud sample

Cloud information for well known clouds can can be obtained through the following command:

```bash
az cloud show -n <cloud name> -o json
```

This is the minimal file attributes that we use in this tool:

```json
{
  "endpoints": {
    "activeDirectory": "https://login.microsoftonline.us",
    "activeDirectoryResourceId": "https://management.core.usgovcloudapi.net/",
    "resourceManager": "https://management.usgovcloudapi.net/"
  }
}
```