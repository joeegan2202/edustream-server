apt update -y
apt install -y nfs-common docker.io
mkdir -p /mnt/nfs
mount -o noac,nconnect=16 10.116.0.8:/mnt/volume_nyc1_01 /mnt/nfs
docker build -t edustream-server .
docker run --publish 80:80 --detach -it \
    -v /mnt/nfs:/nfs \
    --restart always \
    --name run-edustream-server edustream-server
