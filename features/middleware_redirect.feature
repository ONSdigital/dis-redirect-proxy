Feature: Redirect middleware functionality
  As a user of the proxy service
  I want to ensure that the redirect middleware handles URL redirection properly based on Redis settings.

  Scenario: Redirect when Redis contains a valid redirect URL
    Given redis is healthy
    And the feature flag EnableRedisRedirect is set to "true"
    And the redirect proxy is running
    And the key "/old-url" is already set to a value of "/new-url" in the Redis store
    When the Proxy receives a GET request for "/old-url"
      """
      Mock request body
      """
    Then the HTTP status code should be "308"
    And the Location should be "/new-url"

  Scenario: No redirect when Redis contains no redirect URL
    Given redis is healthy
    And the feature flag EnableRedisRedirect is set to "true"
    And the redirect proxy is running
    And redis contains no value for key "/non-redirect-url"
    And the Proxied Service will send the following response with status "200":
      """
      Mock response
      """
    When the Proxy receives a GET request for "/non-redirect-url"
      """
      Mock request body
      """
    Then the HTTP status code should be "200"

  Scenario: No redirect when EnableRedisRedirect is false
    Given redis is healthy
    And the feature flag EnableRedisRedirect is set to "false"
    And the redirect proxy is running
    And the Proxied Service will send the following response with status "200":
      """
      Mock response
      """
    When the Proxy receives a GET request for "/old-url"
      """
      Mock request body
      """
    Then the HTTP status code should be "200"
