# use a slim base image
FROM python:3.8-slim

# set the working directory
WORKDIR /app

# install git and golang
RUN apt-get update && apt-get install -y git golang

# clone go-recon repository
RUN git clone https://github.com/D3Ext/go-recon

# change working directory to repository path
WORKDIR /app/go-recon

# compile and install go-recon
RUN make
RUN make install


