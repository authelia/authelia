Feature: User is redirected to authelia when he is not authenticated

  Scenario: User is redirected to authelia
    Given I'm on https://home.test.local:8080
    When I click on the link to secret.test.local
    Then I'm redirected to "https://auth.test.local:8080/"

