Feature: Generic tests on Authelia endpoints

  Scenario: /api/verify replies with error when redirect parameter is not provided
    When I query "https://authelia.example.com:8080/api/verify"
    Then I get error code 401

  Scenario: /api/verify redirects when redirect parameter is provided
    When I query "https://authelia.example.com:8080/api/verify?redirect=http://login.example.com:8080"
    Then I get redirected to "http://login.example.com:8080"