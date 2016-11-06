# GATS with Netman

To run the [Garden Integration Tests](https://github.com/cloudfoundry/garden-integration-tests)
on your local bosh-lite:

```bash
bosh target lite
bosh update cloud-config bosh-lite/cloud-config.yml
bosh -d bosh-lite/gats.yml deploy
bosh vms # grab the IP address of the garden VM
export GARDEN_ADDRESS=$GARDEN_IP:7777

cd ~/
go get code.cloudfoundry.org/garden-integration-tests
cd $GOPATH/src/code.cloudfoundry.org/garden-integration-tests
go get -t ./...
ginkgo -nodes=4
```
