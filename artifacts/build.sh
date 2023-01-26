chmod +x watchdog 
docker build --no-cache -t rounak316/watchdog:0.0.3 .
# docker tag rounak316/watchdog:0.0.3 rounak316/watchdog:0.0.3
docker push rounak316/watchdog:0.0.3

