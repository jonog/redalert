FROM centurylink/ca-certs
EXPOSE 8888
COPY redalert /
ENTRYPOINT ["/redalert", "server"]
