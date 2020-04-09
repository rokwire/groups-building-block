FROM golang:latest 
RUN mkdir /groups-app
WORKDIR /groups-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make
CMD ["./bin/groups"]