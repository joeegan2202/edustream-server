apt update -y
apt install -y docker.io
mkdir -p /stream
docker build -t edustream-server .
docker run --publish 443:443 --detach -it \
    -v /root/edustream-server:/go/src/app \
    -v /stream:/nfs \
    --restart always \
    --name run-edustream-server edustream-server
