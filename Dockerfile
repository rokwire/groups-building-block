FROM golang:1.21-bullseye as builder

ENV CGO_ENABLED=0

RUN mkdir /groups-app
WORKDIR /groups-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.17

COPY --from=builder /groups-app/bin/groups /
COPY --from=builder /groups-app/docs/swagger.yaml /docs/swagger.yaml

COPY --from=builder /groups-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /groups-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv

COPY --from=builder /groups-app/driver/web/permissions_authorization_policy.csv /driver/web/permissions_authorization_policy.csv
COPY --from=builder /groups-app/driver/web/scope_authorization_policy.csv /driver/web/scope_authorization_policy.csv

COPY --from=builder /groups-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_scope.conf /groups-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_scope.conf
COPY --from=builder /groups-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_string.conf /groups-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_string.conf

COPY --from=builder /etc/passwd /etc/passwd

ENTRYPOINT ["/groups"]