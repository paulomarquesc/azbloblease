#!/bin/bash

# Variables
RESOURCEGROUP_NAME="<resource group name>"
ACCOUNT_NAME="<storage account name>"
CONTAINER_NAME="azbloblease"
LEASE_DURATION=60
SUBSCRIPTION_ID="<subscription id>"
RENEW_ITERATIONS=60
RENEW_ITERATION_WAIT_TIME=30

# Functions
function log()
{
    local MESSAGE=${1}
    echo "$(echo $(date +"%D %T %Z")) - ${MESSAGE}" >> ./${0}.log
}

# Create container and blob to lease
BLOB_CONTAINER_NAME=$(hostname)
../azbloblease/azbloblease createleaseblob \
    -accountname "${ACCOUNT_NAME}" \
    -container "${BLOB_CONTAINER_NAME,,}" \
    -blobname "${BLOB_CONTAINER_NAME}" \
    -resourcegroupname "${RESOURCEGROUP_NAME}" \
    -subscriptionid "${SUBSCRIPTION_ID}"

# Try to acquire lease and become leader
RESULT=$(../azbloblease/azbloblease acquire \
    -accountname $ACCOUNT_NAME \
    -container ${BLOB_CONTAINER_NAME,,} \
    -blobname $BLOB_CONTAINER_NAME \
    -leaseduration $LEASE_DURATION \
    -resourcegroupname $RESOURCEGROUP_NAME \
    -subscriptionid $SUBSCRIPTION_ID 2>/dev/null)

LEASE_ID=$(echo $RESULT | jq -r ".leaseId")
if [[ "${LEASE_ID}" != null ]]; then
    # Running background process to keep leader role - 1 hr 1min
    ../azbloblease/azbloblease renew \
        -accountname $ACCOUNT_NAME \
        -container "${BLOB_CONTAINER_NAME,,}" \
        -blobname $BLOB_CONTAINER_NAME \
        -leaseid $LEASE_ID \
        -resourcegroupname $RESOURCEGROUP_NAME \
        -subscriptionid $SUBSCRIPTION_ID \
        -iterations $RENEW_ITERATIONS \
        -waittimesec $RENEW_ITERATION_WAIT_TIME 2>/dev/null &

    # Real work as leader
    log "I'm the leader (LEASID: ${LEASE_ID}), doing stuff..." 
    
    # Waiting random time up to 30 minutes
    sleep $(( 1 + RANDOM % 1800 ))
else
    log "Could not obtain lease, therefore not being leader, exting"
fi

log "End of script"
