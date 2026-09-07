FROM golang:1.26.5-bookworm
COPY install-deps.sh /tmp/install-deps.sh
RUN sed -i 's/\r$//' /tmp/install-deps.sh && sh /tmp/install-deps.sh
ENV CGO_ENABLED=1 GOTOOLCHAIN=local
WORKDIR /source
ENTRYPOINT ["python3", "/source/BuildTools/build.py"]
