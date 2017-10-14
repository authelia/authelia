Feature: User and groups headers are correctly forwarded to backend
  @need-authenticated-user-john
  Scenario: Custom-Forwarded-User and Custom-Forwarded-Groups are correctly forwarded to protected backend
    When I visit "https://public.test.local:8080/headers"
    Then I see header "Custom-Forwarded-User" set to "john"
    Then I see header "Custom-Forwarded-Groups" set to "dev,admin"
