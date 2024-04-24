# Groups Building Block
The Groups Building Block manages user groups for the Rokwire platform.

## Documentation
The functionality provided by this application is documented in the [Wiki](https://github.com/rokwire/groups-building-block/wiki).

The API documentation is available here: https://api.rokwire.illinois.edu/gr/doc/ui/index.html

## Set Up

### Prerequisites
MongoDB v4.2.2+

Go v1.22+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Format|Required|Description
---|---|---|---
CORE_BB_HOST | < url > | yes | Core BB host URL
CALENDAR_BASE_URL | < url > | yes | Calendar BB host URL
INTERNAL_API_KEY | < string > | yes | Internal API key for invocation by other BBs
GR_MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
GR_MONGO_DATABASE | < string > | yes | MongoDB database name
GR_MONGO_TIMEOUT | < int > | no | MongoDB timeout in milliseconds. Defaults to 500.
NOTIFICATIONS_REPORT_ABUSE_EMAIL | < email > | yes | Email address to send abuse reports to
NOTIFICATIONS_INTERNAL_API_KEY | < string > | yes | Internal API key to use when making requests to the Notifications BB
NOTIFICATIONS_BASE_URL | < url > | yes | URL where the Notifications BB is being hosted
AUTHMAN_BASE_URL | < url > | yes | URL where AuthMan is being hosted
AUTHMAN_USERNAME | < string > | yes | Username to use when logging into to AuthMan
AUTHMAN_PASSWORD | < string > | yes | Password to use when logging into to AuthMan
GROUP_SERVICE_URL | < url > | yes | URL where this application is being hosted
GR_HOST | < url > | yes | URL where this application is being hosted
GR_PORT | < int > | yes | Port where this application is exposed
GR_OIDC_PROVIDER | < url > | yes | URL of OIDC provider to be used when authenticating requests
GR_OIDC_CLIENT_ID | < string > | yes | Client ID to validate with OIDC provider for standard client
GR_OIDC_EXTENDED_CLIENT_IDS | < string > | no | Client ID to validate with OIDC provider for additional clients
GR_OIDC_ADMIN_CLIENT_ID | < url > | yes | Client ID to validate with OIDC for admin client
GR_OIDC_ADMIN_WEB_CLIENT_ID | < url > | yes | Client ID to validate with OIDC for web client
ROKWIRE_API_KEYS | < string (comma-separated) > | yes | List of API keys to be used for client verification
AUTHMAN_ADMIN_UIN_LIST | < string (comma-separated) > | yes | List of UINs for admin users used when loading data from AuthMan
GR_SERVICE_ACCOUNT_ID | < string > | yes | ID of Service Account for Groups BB
GR_PRIV_KEY | < string > | yes | PEM encoded private key for Groups BB

### Run Application

#### Run locally without Docker

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Make the project  
```
$ make
...
▶ building executable(s)… 1.9.0 2020-08-13T10:00:00+0300
```

4. Run the executable
```
$ ./bin/notifications
```

#### Run locally as Docker container

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Create Docker image  
```
docker build -t notifications .
```
4. Run as Docker container
```
docker-compose up
```

#### Tools

##### Run tests
```
$ make tests
```

##### Run code coverage tests
```
$ make cover
```

##### Run golint
```
$ make lint
```

##### Run gofmt to check formatting on all source files
```
$ make checkfmt
```

##### Run gofmt to fix formatting on all source files
```
$ make fixfmt
```

##### Cleanup everything
```
$ make clean
```

##### Run help
```
$ make help
```

##### Generate Swagger docs
```
$ make swagger
```

### Test Application APIs

Verify the service is running as calling the get version API.

#### Call get version API

curl -X GET -i https://api-dev.rokwire.illinois.edu/gr/version

Response
```
0.1.2
```

## Contributing
If you would like to contribute to this project, please be sure to read the [Contributing Guidelines](CONTRIBUTING.md), [Code of Conduct](CODE_OF_CONDUCT.md), and [Conventions](CONVENTIONS.md) before beginning.

### Secret Detection
This repository is configured with a [pre-commit](https://pre-commit.com/) hook that runs [Yelp's Detect Secrets](https://github.com/Yelp/detect-secrets). If you intend to contribute directly to this repository, you must install pre-commit on your local machine to ensure that no secrets are pushed accidentally.

```
# Install software 
$ git pull  # Pull in pre-commit configuration & baseline 
$ pip install pre-commit 
$ pre-commit install
```