# FROM envoyproxy/envoy-dev:7b0ce0d32a9b584626e8c16b5ae07817eade322d
# COPY envoysimple.yaml /etc/envoy/envoy.yaml
FROM envoyproxy/envoy-dev:latest

RUN apt-get update
COPY envoy.yaml /etc/
CMD /usr/local/bin/envoy -c /etc/envoy.yaml --log-level trace
EXPOSE 9901
EXPOSE 9090