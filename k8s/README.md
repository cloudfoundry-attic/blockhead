This directory has the resources to install the broker in k8s. Only need to be
used once on startup.

`rbac.yaml` sets up the default namespace so that the broker has the access
needed to create k8s objects.

`deploy.yaml` installs the broker and creates a service to access it in-cluster.

`ingress.yaml` sets up the basic ingress for accessing the broker from outside
the cluster by a resolving dns name. It needs to be edited to have the right
host as the ingress domain.

`n.yaml` has test resources to troubleshoot ingress or other networking with a
very basic nginx setup.

* How to K8s

Prereqs:
 - get a kube cluster
 - know an address to access it (it's load-balancer or worker-node)
 
RBAC setup. We only require access to create and delete services and pods. The
provided rbac has much more access because I am lazy.

create it with :

    kubectl create -f rbac.yaml
    
Edit the `deploy.yaml` to fixup the `external_address` field so that it points
at either the cluster loadbalancer or a publicly accessible node address (IP or DNS).
Now create the Broker, it's Service, and the config data necessary to start it with:

    kubectl create -f deploy.yaml

Wait a bit and everything should be deployed.

    kubectl get all -l app=blockhead-broker

In that should be the service and what nodeport it is bound to.

We can get the catalog with curl.

```
curl -sSL  a:b@peanuts.sng01.containers.appdomain.cloud:32089/v2/catalog  -H
"X-Broker-API-Version: 2.14"

{"services":[{"id":"0d25f970-899a-4aa4-b753-640e33b66389","name":"eth","description":"Ethereum
Geth
Node","bindable":true,"tags":["eth","geth","dev"],"plan_updateable":false,"plans":[{"id":"0a20f765-9e22-4ac2-b9f7-2ac8f706747c","name":"free","description":"Free
Trial","free":true}],"metadata":{"displayName":"Geth 1.8"}}]}
```

Do a service provision:

```console
$ curl -sSL  a:b@peanuts.sng01.containers.appdomain.cloud:30000/v2/service_instances/test-broker -d "@sp.json" -XPUT -H "X-Broker-API-Version: 2.14"   -H "Content-Type: application/json"
```
where `sp.json` contains the service and plan to provision.


Do a bind:

```console
$ curl http://a:b@peanuts.sng01.containers.appdomain.cloud:30000/v2/service_instances/test-broker/service_bindings/newbind -d "@sp2.json" -XPUT -H "X-Broker-API-Version: 2.14"   -H "Content-Type: application/json"
{"credentials":{"ContainerInfo":{"ExternalAddress":"peanuts.sng01.containers.appdomain.cloud","InternalAddress":"test-broker","Bindings":{"8545":[{"Port":"30971"}]}},"NodeInfo":{"address":"0x6361eFC13f0b5f0072e3774254a1e9a876db6BD7","abi":"[{\"constant\":false,\"inputs\":[],\"name\":\"voteA\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"proposalA\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"voteB\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"proposalB\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]","contract_address":"0xA5948eDC459d0D3e83D6f9b9Cf089182EBE2b056","gas_price":"1","transaction_hash":"0xfa6d497809439d305cd64073de44c8e4b51ebc9d9a24fc1a651bd73a6f0b85c7"}}}
```
Where `sp2.json` contains the service, plan, and parameters.

If attaching remix from the webconsole, make sure that it's loaded from the insecure http backend.

* cleanup tips

Delete all the deployed geth nodes in one shot

    kubectl delete all -l provisionedBy=blockhead-broker
    
Delete all the broker stuff in one shot

    kubectl delete all -l app=blockhead-broker

* optional stuff

using ingress is nice. Easiest to be used for the broker. The host role in `ingress.yaml` has to be adjusted.

    kubectl create -f ingress.yaml

Can of course deploy the app in kubernetes.


* set up an IKS cluster

On IKS, your loadbalancer DNS entry is found in the output of:

```console
$ bx cs cluster-get  mhb-blockhead-broker
Retrieving cluster mhb-blockhead-broker...
OK

                           
Name:                   mhb-blockhead-broker   
ID:                     26760c5d3aca48619e4b79587feb1ce2   
State:                  normal   
Created:                2018-09-29T01:47:10+0000   
Location:               sao01   
Master URL:             https://169.57.151.10:31423   
Master Location:        sao01   
Master Status:          Ready (2 days ago)   
Ingress Subdomain:      mhb-blockhead-broker.sao01.containers.appdomain.cloud   
Ingress Secret:         mhb-blockhead-broker   
Workers:                1   
Worker Zones:           sao01   
Version:                1.11.3_1524   
Owner:                  mbauer@us.ibm.com   
Monitoring Dashboard:   -   
```

Set a region if necessary:

     bx cs region-set ap-north

Look for *Ingress Subdomain*, ours is `mhb-blockhead-broker.sao01.containers.appdomain.cloud`

```console
$ bx cs machine-types sng01
OK
Name        Cores   Memory   Network Speed   OS             Server Type   Storage   Secondary Storage   Trustable   
u2c.2x4     2       4GB      1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
c2c.16x16   16      16GB     1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
c2c.16x32   16      32GB     1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
b2c.4x16    4       16GB     1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
b2c.8x32    8       32GB     1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
b2c.16x64   16      64GB     1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
```


look up the vlans to add

```
$ bx cs vlans --zone sng01
OK
ID        Name   Number   Type      Router         Supports Virtual Workers   
2457919          1458     private   bcr01a.sng01   true   
2457917          1406     public    fcr01a.sng01   true   
```

```
$ bx cs cluster-create --name peanuts --kube-version 1.11.3 --location sng01  --machine-type u2c.2x4 --workers 3  --public-vlan 2457917 --private-vlan 2457919
Creating cluster...
OK
```

Eventually...

```
$ bx cs clusters
OK
Name      ID                                 State       Created          Workers   Location   Version   
peanuts   ed40c720877447e192fb51e09f7e4474   requested   30 seconds ago   1         sng01      1.11.3_1524   
```
