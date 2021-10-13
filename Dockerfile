FROM golang:1.16-buster as builder

ENV CGO_ENABLED=0

RUN mkdir /groups-app
WORKDIR /groups-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.11.6

COPY --from=builder /groups-app/bin/groups /
COPY --from=builder /groups-app/docs/swagger.yaml /docs/swagger.yaml

COPY --from=builder /groups-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /groups-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv
COPY --from=builder /groups-app/driver/web/permissions_authorization_policy.csv /driver/web/permissions_authorization_policy.csv
COPY --from=builder /groups-app/driver/web/scope_authorization_policy.csv /driver/web/scope_authorization_policy.csv

COPY --from=builder /etc/passwd /etc/passwd

ENTRYPOINT ["/groups"]