@Healthcheck
Feature: Healthcheck endpoint should inform the health of service

    Scenario: Returning a OK (200) status when health endpoint called
        Given redis is healthy
        And the redirect proxy is running
        And I wait 2 seconds for the healthcheck to be available
        When I GET "/health"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And the health checks should have completed within 2 seconds
        And I should receive the following health JSON response:
        """
            {
              "status": "OK",
              "version": {
                "git_commit": "132a3b8570fdfc9098757d841c8c058ddbd1c8fc",
                "language": "go",
                "language_version": "go1.24.2",
                "version": "v1.2.3"
              },
              "checks": [
                {
                  "name": "Redis",
                  "status": "OK",
                  "status_code": 200,
                  "message": "redis is healthy"
                }
              ]
            }
        """

    Scenario: Returning a WARNING (429) status when health endpoint called
        Given redis stops running
        And I have a healthcheck interval of 1 second
        And I wait 2 seconds for the healthcheck to be available
        When I GET "/health"
        Then the HTTP status code should be "429"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And the health checks should have completed within 4 seconds
        And I should receive the following health JSON response:
        """
            {
                "status": "WARNING",
                "version": {
                    "git_commit": "132a3b8570fdfc9098757d841c8c058ddbd1c8fc",
                    "language": "go",
                    "language_version": "go1.17.8",
                    "version": "v1.2.3"
                },
                "checks": [
                    {
                        "name": "Redis",
                        "status": "CRITICAL",
                        "status_code": 500,
                        "message": "dial tcp 127.0.0.1:6379: connect: connection refused"
                    }
                ]
            }
        """

    Scenario: Returning a CRITICAL (500) status when health endpoint called
        Given redis stops running
        And I have a healthcheck interval of 1 second
        And I wait 6 seconds to pass the critical timeout
        When I GET "/health"
        Then the HTTP status code should be "500"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And the health checks should have completed within 6 seconds
        And I should receive the following health JSON response:
        """
            {
                "status": "CRITICAL",
                "version": {
                    "git_commit": "132a3b8570fdfc9098757d841c8c058ddbd1c8fc",
                    "language": "go",
                    "language_version": "go1.17.8",
                    "version": "v1.2.3"
                },
                "checks": [
                    {
                        "name": "Redis",
                        "status": "CRITICAL",
                        "status_code": 500,
                        "message": "dial tcp 127.0.0.1:6379: connect: connection refused"
                    }
                ]
            }
        """
