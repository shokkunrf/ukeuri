FROM golang:1.16.3-buster

ARG USERNAME=gopher
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd -s /bin/bash  --uid $USER_UID --gid $USER_GID -m $USERNAME 
