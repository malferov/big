## deploy to local environment
create directories
```
mkdir data data/docker_mirror_cache data/docker_mirror_certs
```
generate ssl certificate
```
openssl req -newkey rsa -keyout data/big.key -x509 -nodes -days 3650 -out data/big.crt
# ubuntu
sudo cp data/big.crt /usr/local/share/ca-certificates/big.crt
sudo update-ca-certificates
# centos
sudo cp data/big.crt /etc/pki/ca-trust/source/anchors
sudo update-ca-trust
```
run nginx on localhost with the following config
```
sudo cp big.conf /etc/nginx/conf.d
```
deploy locally
```
docker-compose up -d
echo "127.0.0.1 r.big.com" > /etc/hosts
```
tests
```
curl https://r.big.com/version -v | jq
docker pull r.big.com/123456/<app>:<tag>
```
