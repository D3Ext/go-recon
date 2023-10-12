FROM golang:1.19

# set work directory
WORKDIR /app

# update packages and install git
RUN apt-get update && apt-get install -y git

# clone go-recon repository
RUN git clone https://github.com/D3Ext/go-recon

# change working directory to repository path
WORKDIR /app/go-recon

# compile and install go-recon
RUN make
RUN make install


