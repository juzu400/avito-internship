FROM ubuntu:latest
LABEL authors="catsc"

ENTRYPOINT ["top", "-b"]