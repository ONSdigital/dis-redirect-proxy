Feature: Healthcheck endpoint should inform the health of service

    Scenario: Returning a OK (200) status when health endpoint called
        Given the redirect proxy is initialised
        And redis is healthy
        And I run the redirect proxy
        And I wait 2 seconds for the healthcheck to be available
#        When I GET "/health"
#        Then the HTTP status code should be "200"
#        And the response header "Content-Type" should be "application/json; charset=utf-8"
#        And I should receive the following health JSON response:
        """
            {
              "status": "OK",
              "version": {
                "git_commit": "132a3b8570fdfc9098757d841c8c058ddbd1c8fc",
                "language": "go",
                "language_version": "go1.24.2",
                "version": ""
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
