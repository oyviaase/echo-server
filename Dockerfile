FROM scratch
COPY bin/echo-server /bin/echo-server
ENV PORT 8080
EXPOSE 8080
ENV ADD_HEADERS='{"X-Real-Server": "echo-server"}'
ENTRYPOINT ["/bin/echo-server"]
