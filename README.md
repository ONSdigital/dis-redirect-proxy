# dis-redirect-proxy

A Go Service to redirect legacy URLs requested by users

## Getting started

* Run `make debug` to run application on http://localhost:30000
* Run `make help` to see full list of make targets

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                  | Description                                                                                                        |
|------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :30000                   | The host and port to bind to                                                                                       |
| ENABLE_REDIRECTS             | false                    | Feature flag to enable middleware redis check for redirects                                                        |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| OTEL_EXPORTER_OTLP_ENDPOINT  | localhost:4317           | Endpoint for OpenTelemetry service                                                                                 |
| OTEL_SERVICE_NAME            | dis-redirect-proxy       | Label of service for OpenTelemetry service                                                                         |
| OTEL_BATCH_TIMEOUT           | 5s                       | Timeout for OpenTelemetry                                                                                          |
| OTEL_ENABLED                 | false                    | Feature flag to enable OpenTelemetry                                                                               |
| PROXIED_SERVICE_URL          | <http://localhost:20000> | The service address where requests are forwarded to by default                                                     |
| REDIS_ADDRESS                | localhost:6379           | Endpoint for Redis service                                                                                         |

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
