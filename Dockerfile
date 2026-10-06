FROM gcr.io/distroless/static-debian12:nonroot
EXPOSE 8888
COPY redalert /
ENTRYPOINT ["/redalert", "server"]
