@ReleaseFallback
Feature: Release fallback functionality
  As a user of the proxy service
  I want to ensure that the fallback to Wagtail for /releases/ URLs works correctly.

  Scenario: Get from proxy with EnableReleasesFallback false
    Given the Proxied Service will send the following response with status "200":
      """
      Mock response
      """
    And the feature flag EnableReleasesFallback is set to "false"
    When the Proxy receives a GET request for "/releases/some-release"
      """
      Mock request body
      """
    Then the response from the Proxied Service should be returned unmodified by the Proxy

  Scenario: Get from proxy with EnableReleasesFallback true
    Given the Proxied Service will send the following response with status "200":
      """
      Mock response
      """
    And the feature flag EnableReleasesFallback is set to "true"
    When the Proxy receives a GET request for "/bulletins/my-bulletin"
      """
      Mock request body
      """
    Then the response from the Proxied Service should be returned unmodified by the Proxy

  Scenario: Get release from Wagtail
    Given the Wagtail Service will send the following response with status "200":
      """
      Mock response
      """
    And the feature flag EnableReleasesFallback is set to "true"
    And the Wagtail Service will set the "ETag" header to "abc123"
    And the Wagtail Service will set the "Referrer-Policy" header to "origin"
    When the Proxy receives a <request-method> request for "/releases/some-release"
      """
      Mock request body
      """
    Then the response from the Wagtail Service should be returned unmodified by the Proxy
    Examples:
      | request-method |
      | GET            |
      | POST           |
      | PUT            |
      | PATCH          |
      | DELETE         |

  Scenario: Get release from proxy via Wagtail
    Given the Wagtail Service will send the following response with status "404":
      """
      Mock Wagtail response
      """
    And the Proxied Service will send the following response with status "200":
      """
      Mock proxied service response
      """
    And the feature flag EnableReleasesFallback is set to "true"
    When the Proxy receives a <request-method> request for "/releases/some-release"
      """
      Mock request body
      """
    Then the response from the Proxied Service should be returned unmodified by the Proxy
    Examples:
      | request-method |
      | GET            |
      | POST           |
      | PUT            |
      | PATCH          |
      | DELETE         |

  Scenario: Get release from Wagtail - non 404 codes
    Given the Wagtail Service will send the following response with status "<status_code>":
      """
      Mock Wagtail response
      """
    And the feature flag EnableReleasesFallback is set to "true"
    When the Proxy receives a GET request for "/releases/some-release"
      """
      Mock request body
      """
    Then the response from the Wagtail Service should be returned unmodified by the Proxy
    Examples:
      | status_code |
      | 200         |
      | 301         |
      | 302         |
      | 307         |
      | 308         |
      | 400         |
      | 401         |
      | 403         |
      | 410         |
      | 500         |
      | 502         |
