FROM scratch
ARG TARGETPLATFORM
COPY artifacts/build/release/$TARGETPLATFORM/echo-server /bin/echo-server
COPY bin/cert.pem /bin/cert.pem
COPY bin/key.pem /bin/key.pem
ENV PORT 8080
ENV SSLPORT 8443

EXPOSE 8080 8443

ENV ADD_HEADERS='{"X-Real-Server": "echo-server"}'

WORKDIR /bin
ENTRYPOINT ["/bin/echo-server"]
