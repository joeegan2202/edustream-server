apt update -y
apt install -y docker.io
mkdir -p /stream
docker build -t edustream-server .
docker run --publish 443:443 --detach -it \
    -v /stream:/nfs \
    --restart always \
    --name run-edustream-server edustream-server
