FROM scratch
COPY goMqttModbus /goMqttModbus
EXPOSE 8080
ENTRYPOINT ["/goMqttModbus"]
CMD ["-port", "8080", "-config", "/data/config.json", "-log", "info"]
