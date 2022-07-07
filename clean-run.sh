#! /bin/sh

cd ./frontend

npm install
npm run build

cd ..
docker-compose up --force-recreate --build -d
docker image prune -f

docker-compose up