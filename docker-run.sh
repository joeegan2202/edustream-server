docker build -t edustream-server .
docker run --publish 80:80 --detach -it --rm --name run-edustream-server edustream-server
