# sudo docker build . -t gateway_tp:1.0
FROM ubuntu:trusty
# Install GO and other stuff that is needed
RUN apt-get update
RUN apt-get install wget xterm -y
RUN cd /tmp
RUN wget https://storage.googleapis.com/golang/go1.14.7.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.14.7.linux-amd64.tar.gz
RUN apt-get install -y libpcap-dev
RUN apt-get install -y libc6 libc6-dev
RUN apt-get install -y curl
RUN apt-get install -y tcpdump
# Set up the env variables
ENV GOPATH=$HOME/work
ENV GOBIN=$GOPATH/bin
ENV PATH=$PATH:/usr/local/go/bin
# Make the GO directories
RUN ["mkdir", "-p", "$GOPATH"] # mkdir -p $GOPATH
RUN ["mkdir", "-p", "$GOBIN"] #mkdir -p $GOBIN
# Make the directory for the Gateway TP
RUN mkdir /Gateway_TP
# Copy the codebase into that folder
COPY ./ /Gateway_TP/
# At runtime set the configuration, build main.go and run it
