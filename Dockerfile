FROM golang AS build
ARG version
ARG commit
ARG date
WORKDIR /go/src
COPY src/ .
RUN go get github.com/gin-gonic/gin
RUN go get github.com/golang/glog
RUN go build -o proxy -ldflags "-X main.version=$version -X main.commit=$commit -X 'main.date=$date'"

FROM centos
EXPOSE 5001
COPY --from=build /go/src/proxy .
ENTRYPOINT ["./proxy"]
CMD ["-stderrthreshold=ERROR"]
