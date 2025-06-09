Feature: Unmodified response

  Under normal circumstances, the proxy will return the ProxiedService response to the client completely unmodified. This
  means that the body, headers and status code will be exactly the same and will remain unchanged.

  Scenario Outline: The request method is not GET or HEAD
    Given the Proxied Service will send the following response with status "200":
      """
      Mock response
      """
    And the Proxied Service will set the "ETag" header to "abc123"
    And the Proxied Service will set the "Referrer-Policy" header to "origin"
    When the Proxy receives a <request-method> request for "/"
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
