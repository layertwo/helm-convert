FROM golang:1.22 AS build
ENV GOPATH=""
ARG LDFLAGS
COPY go.sum .
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o helm-convert -ldflags "$LDFLAGS" main.go

FROM alpine:3.19 AS helm
ENV HELM_BASE_URL=https://get.helm.sh
ENV HELM_VERSION=v3.14.4
ENV HELM_TMP_FILE=helm-${HELM_VERSION}-linux-amd64.tar.gz
RUN wget ${HELM_BASE_URL}/${HELM_TMP_FILE} && wget ${HELM_BASE_URL}/${HELM_TMP_FILE}.sha256sum
RUN sha256sum -c ${HELM_TMP_FILE}.sha256sum
RUN tar -xvf helm-${HELM_VERSION}-linux-amd64.tar.gz

FROM alpine:3.19
COPY --from=helm /linux-amd64/helm /usr/local/bin/helm
RUN mkdir -p /root/.helm/plugins/helm-convert
COPY plugin.yaml /root/.helm/plugins/helm-convert/plugin.yaml
COPY --from=build /go/helm-convert /root/.helm/plugins/helm-convert/helm-convert
ENTRYPOINT ["helm"]
