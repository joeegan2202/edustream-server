apt update -y
apt install -y docker.io
mkdir -p /stream
docker build -t edustream-server .
docker run --publish 80:80 --detach -it \
    -v /stream:/nfs \
    --restart always \
    --name run-edustream-server edustream-server
