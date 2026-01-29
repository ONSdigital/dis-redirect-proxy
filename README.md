# dis-redirect-proxy

A Go Service to redirect legacy URLs requested by users.

It also has fallback functionality for particular paths where it is not known which service contains which URL.

## Getting started

* Run `make debug` to run application on <http://localhost:30000>
* Run `make help` to see full list of make targets

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Tools

To run some of our tests you will need additional tooling:

#### Audit

We use `dis-vulncheck` for auditing, which you will [need to install](https://github.com/ONSdigital/dis-vulncheck).

### Configuration

| Environment variable         | Default                  | Description                                                                                                        |
|------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :30000                   | The host and port to bind to                                                                                       |
| ENABLE_REDIRECTS             | false                    | Feature flag to enable middleware redis check for redirects                                                        |
| ENABLE_RELEASES_FALLBACK     | false                    | Enable fallback routing for /releases/                                                                             |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| PROXIED_SERVICE_URL          | <http://localhost:20000> | The service address where requests are forwarded to by default                                                     |
| OTEL_EXPORTER_OTLP_ENDPOINT  | localhost:4317           | Endpoint for OpenTelemetry service                                                                                 |
| OTEL_SERVICE_NAME            | dis-redirect-proxy       | Label of service for OpenTelemetry service                                                                         |
| OTEL_BATCH_TIMEOUT           | 5s                       | Timeout for OpenTelemetry                                                                                          |
| OTEL_ENABLED                 | false                    | Feature flag to enable OpenTelemetry                                                                               |
| REDIS_ADDRESS                | localhost:6379           | Endpoint for Redis service                                                                                         |
| REDIRECT_API_URL             | localhost:29900          | Currently used to populated HATEOS links                                                                           |
| REDIS_ADDRESS                | localhost:6379           | Endpoint for Redis service                                                                                         |
| REDIS_CLUSTER_NAME           | ""                       | Cluster name for Redis service                                                                                     |
| REDIS_REGION                 | ""                       | AWS Region to connect to for Redis backing service                                                                 |
| REDIS_SEC_PROTO              | ""                       | Use 'TLS' to connect with TLS                                                                                      |
| REDIS_SERVICE                | ""                       | Name of the redis service to connect to, e.g. memorydb, elasticache                                                |
| REDIS_USERNAME               | ""                       | Username to connect to Redis with                                                                                  |
| WAGTAIL_URL                  | <http://localhost:8000>  | URL for Wagtail - this shouldn't be so specific but it's a fairly specific piece of functionality                  |

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
