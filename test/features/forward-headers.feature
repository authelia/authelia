Feature: Headers are correctly forwarded to backend
  @need-authenticated-user-john
  Scenario: Custom-Forwarded-User and Custom-Forwarded-Groups are correctly forwarded to protected backend
    When I visit "https://public.example.com:8080/headers"
    Then I see header "Custom-Forwarded-User" set to "john"
    Then I see header "Custom-Forwarded-Groups" set to "dev,admin"

  Scenario: Custom-Forwarded-User and Custom-Forwarded-Groups are correctly forwarded to protected backend when basic auth is used
    When I request "https://single_factor.example.com:8080/headers" with username "john" and password "password" using basic authentication
    Then I received header "Custom-Forwarded-User" set to "john"
    And I received header "Custom-Forwarded-Groups" set to "dev,admin"