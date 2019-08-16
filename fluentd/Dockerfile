# fluentd/Dockerfile
#FROM fluent/fluentd:v0.12
#RUN ["gem", "install", "fluent-plugin-elasticsearch", "--no-rdoc", "--no-ri", "--version", "1.18.0"]
#RUN ["gem", "install", "fluent-plugin-elasticsearch", "--no-rdoc", "--no-ri"]
# FROM fluent/fluentd:v1.6-debian-1

FROM fluent/fluentd:v1.6-1

# Use root account to use apk
USER root

# below RUN includes plugin as examples elasticsearch is not required
# you may customize including plugins as you wish
RUN apk add --no-cache --update --virtual .build-deps \
    sudo build-base ruby-dev \
    && sudo gem install fluent-plugin-elasticsearch \
    && sudo gem sources --clear-all \
    && apk del .build-deps \
    && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

COPY conf/fluent.conf /fluentd/etc/
#COPY entrypoint.sh /bin/
EXPOSE 24224
USER fluent
CMD exec fluentd -c /fluentd/etc/fluent.conf