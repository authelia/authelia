@needs-regulation-config
Feature: Authelia regulates authentication to avoid brute force

  @need-registered-user-blackhat
  Scenario: Attacker tries too many authentication in a short period of time and get banned
    Given I visit "https://login.example.com:8080/"
    And I set field "username" to "blackhat"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    When I set field "password" to "password"
    And I click on "Sign in"
    Then I get a notification of type "error" with message "Authentication failed. Please check your credentials."

  @need-registered-user-blackhat
  Scenario: User is unbanned after a configured amount of time
    Given I visit "https://login.example.com:8080/?rd=https://public.example.com:8080/secret.html"
    And I set field "username" to "blackhat"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    When I wait 6 seconds 
    And I set field "password" to "password"
    And I click on "Sign in"
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://public.example.com:8080/secret.html"
