FROM golang:latest

WORKDIR /home

COPY main.go /home/main.go

RUN ["go", "get", "github.com/gorilla/mux"]
RUN ["go", "get", "git.darknebu.la/GalaxySimulator/structs"]

ENTRYPOINT ["go", "run", "/home/main.go"]
