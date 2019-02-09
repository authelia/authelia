Feature: Authentication scenarii

  Scenario: Logout redirects user to redirect URL given in parameter
    When I visit "https://login.example.com:8080/logout?rd=https://home.example.com:8080/"
    Then I'm redirected to "https://home.example.com:8080/"
