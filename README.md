# container-registry-proxy

## build proxy application
```
git checkout master
cd src
go test -v
# use your registry where you whant to push the image; run `docker login` if necessary
./build.sh <private_registry>
```

## application deployment
In the `deployment.yml` file fill in the following parameters  
set <private_registry> placeholder for proxy image field  
set `PROXY_APPS` container environment variable based on the following pattern. this variable controls 
routing between requested product and appropriate container registry
```
PROXY_APPS="<app_name>:<app_registry>:<namespace> <app_name2>:<app_registry2>:<namespace2> ..."
```
set gitlab credentials `PROXY_GITLAB="username:access_token"`  
set `AUTH_REGISTRIES` container environment variable for `rpardini` service. 
please find documentation at https://github.com/rpardini/docker-registry-proxy  
you may need json credentials file to access your gcr.io private registry. you can use the following helper command
```
export JSON_KEY=$(cat json.key)
# use $JSON_KEY variable in deployment.yml
envsubst < deployment.yml | kubectl apply -f -
```
deploy application via `kubectl`
```
# ssl certificate for r.bigdataboutique.com domain
kubectl create secret tls proxy-tls --cert=<file.crt> --key=<file.key>
kubectl apply -f deployment.yml # or command above
kubectl apply -f service.yml
kubectl apply -f ingress.yml
```
create `dns` record for r.bigdataboutique.com hostname
```
kubectl get ingress proxy -o=jsonpath="{.status.loadBalancer.ingress[].ip}"
```

## testing
```
# check version
curl https://r.bigdataboutique.com/version -v | jq
# pull the image
docker pull r.bigdataboutique.com/123456/<app>:<tag>
```

## debugging
add `-stderrthreshold=INFO` command line argument for more verbose logging. see `./proxy --help` for details
