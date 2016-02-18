# Kansible

Orchestrate processes with Kubernetes and Ansible for cases where you have not yet dockerized your processes, or your process currently needs to run on Windows, AIX, Solaris or HP-UX or an old Linux distro that predates docker

## Overview

Kansible is a simple tool for orchestrating your non container based processes in the same way as you orchestrate all your containers inside Kubernetes. 

## How it works

You use kansible as follows:

* create an [Ansible playbook](http://docs.ansible.com/ansible/playbooks.html) to _install and provision_ the software you wish to run on a number of machines defined by the [Ansible inventory](http://docs.ansible.com/ansible/intro_inventory.html)
* run the [Ansible playbook](http://docs.ansible.com/ansible/playbooks.html) either as part of a [CI / CD build pipeline](http://fabric8.io/guide/cdelivery.html) when there's a change to the git repo of the Playbook, or using a command line tool, cron or [Ansible Tower](https://www.ansible.com/tower)
* define a Replication Controller YAML file for running the command for your process [like this example](https://github.com/fabric8io/fabric8-ansible-spring-boot/blob/master/rc.yml#L15-L16). Its mostly boilerplate other than the actual command you need to run execute your process. Notice you can use the `{{ foo_bar }}` ansible variable expressions to refer to variables from your [global ansible variables file](https://github.com/fabric8io/fabric8-ansible-spring-boot/blob/master/group_vars/appservers)
* whenever the playbook git repo changes, run the **kansible rc** command inside a clone of the playbook git repository:

    kansible rc myhosts
    
where `myhosts` is the name of the hosts you wish to use in the Ansible inventory.    

Then **kansible** will create a Replication Controller of kansible pods which will start and supervise your processes. 

For each remote process on Windows, Linux, Solaris, AIX, HPUX kansible will create a kansible pod in Kubernetes.

### Using kansible

* As proceses start and stop, you'll see the processes appear or disappear inside kubernetes, the CLI, REST API or the console.
* You can scale up and down the Replication Controller via CLI, REST API or console.
* You can then view the logs of any process in the usual kubernetes way via the command line, REST API or web console. 
* [Centralised logging](http://fabric8.io/guide/logging.html) then works great on all your processes (providing the command you run outputs logs to `stdout` / `stderr`)

### Exposing ports

Any ports defined in the Replication Controller YAML file will be automatically forwarded to the remote process.

This means you can take advantage of things like [centralised metrics and alerting](http://fabric8.io/guide/metrics.html) or Kubernetes Services and the built in service discovery and load balancing inside Kubernetes!

### Opening a shell on the remote process

You can open a shell directly on the remote machine via the web console or by running 

    oc exec -p mypodname bash
        
### Running Kansible
  
To try out running one of the example Ansible provisioned apps try the following:

* [Download a release](https://github.com/fabric8io/kansible/releases) and add `kansible` to your `$PATH` 
* Or [Build kansible](https://github.com/fabric8io/kansible#building) then add the `$PWD/bin` folder to your `$PATH` so that you can type in `kansible` on the command line

The following examples use these files:

* `inventory` is the Ansible inventory file used unless you specify the `--inventory` command line option
* `rc.yml` is the Replication Controller configuration used for the supervisor pods unless you specify the `--rc` command line option

 
#### for [fabric8-ansible-spring-boot](https://github.com/fabric8io/fabric8-ansible-spring-boot)

type the following to setup the VMs and provision things with Ansible

```
    git clone https://github.com/fabric8io/fabric8-ansible-spring-boot.git
    cd fabric8-ansible-spring-boot
    vagrant up
    ansible-playbook -i inventory provisioning/site.yml -vv
```
    
Now to setup the Replication Controller for the supervisors run the following, where `appservers` is the hosts from the inventory
    
```    
    kansible rc appservers
```      

The pods should now start up for each host in the inventory!


### [fabric8-ansible-hawtapp](https://github.com/fabric8io/fabric8-ansible-hawtapp)

type the following to setup the VMs and provision things with Ansible

```
    git clone https://github.com/fabric8io/fabric8-ansible-hawtapp.git
    cd fabric8-ansible-hawtapp
    vagrant up
    ansible-playbook -i inventory provisioning/site.yml -vv
```
    
Now to setup the Replication Controller for the supervisors run the following, where `appservers` is the hosts from the inventory

```    
    kansible rc appservers
```      

The pods should now start up for each host in the inventory!

### Using Windows

To use windows you need to first make sure you've installed **pywinrm**:

    sudo pip install pywinrm
    
To try using windows machines, replace `appservers` with `winboxes` in the above commands; assuming you have created the [Windows vagrant machine](https://github.com/fabric8io/fabric8-ansible-hawtapp/tree/master/windows) locally

Or you can add the windows machine into the `appservers` hosts section in the `inventory` file.


### Configuring kansible

The best way to configure if you want to connect via SSH for unix machines or WinRM for windows machines is via the Ansible Inventory.

By default SSH is used on port 22 unless you specify `ansible_port` in the inventory or specify `--port` on the command line.

You can configure Windows machines using the `ansible_connection=winrm` property in the inventory:


```yaml
[winboxes]
windows1 ansible_host=localhost ansible_port=5985 ansible_user=foo ansible_pass=somepasswd! ansible_connection=winrm

[unixes]
app1 ansible_host=10.10.3.20 ansible_user=vagrant ansible_ssh_private_key_file=.vagrant/machines/app1/virtualbox/private_key
app2 ansible_host=10.10.3.21 ansible_user=vagrant ansible_ssh_private_key_file=.vagrant/machines/app2/virtualbox/private_key
```

You can also enable WinRM via the `--winrm` command line flag: 

```
export KANSIBLE_WINRM=true
kansible pod --winrm somehosts somecommand

```

or by setting the environment variable `KANSIBLE_WINRM` which is a little easier to configure on the RC YAML:

```
export KANSIBLE_WINRM=true
kansible pod somehosts somecommand

```
### Checking the runtime status of the supervisors
 
To see which pods own which hosts run the following command:
 
```
    oc export rc hawtapp-demo | grep ansible.fabric8  | sort
```

Where `hawtapp-demo` is the name of the RC for the supervisors.

The output is of the format:

```
    pod.ansible.fabric8.io/app1: supervisor-znuj5
    pod.ansible.fabric8.io/app2: supervisor-1same
```

Where the output is of the form ` pod.ansible.fabric8.io/$HOSTNAME: $PODNAME`



 
## Building
 
 * install [glide](https://github.com/Masterminds/glide#install)
 * prepare the `vendor` folder by typing
 
 ```
     make bootstrap
     export GO15VENDOREXPERIMENT=1
 ```
     
 * then to build the binary
     
 ```
     make build
 ```
     
 * you can then run it via
 
 ```    
     ./bin/kansible
 ```

## Running pods locally

You can run `kansible rc ...` easily on a local build when working on the code. However to try out changes to the pod for `kansible pod ...` you can run that locally outside of docker with a small trick.

You must set the `HOSTNAME` environment variable to a valid pod name you wish to use.

```
    export HOSTNAME=fabric8-znuj5
```

e.g. the above uses the pod name for the current fabric8 console.

This lets you pretend to be different pods from the command line when trying it out locally. e.g. run the `kansible pod ...` command in 2 shells as different pods, provided the `HOSTNAME` values are diferent.

The pod name must be valid as `kansible pod ...` command will update the pod to annotate which host its chosen etc.

So to run the [above examples](#running-kansible) type the following:

for [fabric8-ansible-spring-boot](https://github.com/fabric8io/fabric8-ansible-spring-boot):
    
```    
    kansible pod -rc hawtapp-demo appservers /opt/cdi-camel-2.2.98-SNAPSHOT-app/bin/run.sh
```      

for [fabric8-ansible-hawtapp](https://github.com/fabric8io/fabric8-ansible-hawtapp):

```    
    kansible pod  -rc springboot-demo appservers /opt/springboot-camel-2.2.98-SNAPSHOT
```      


## License

Copyright 2016 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
