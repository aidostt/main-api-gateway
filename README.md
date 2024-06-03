# reservista


### To run the container you should write
docker build -t reservista .
docker run -p 8000:8000 --env-file .env -ti reservista
