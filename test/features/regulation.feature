Feature: Authelia regulates authentication to avoid brute force

  @needs-test-config
  Scenario: Attacker tries too many authentication in a short period of time and get banned
    Given I visit "https://auth.test.local:8080/"
    And I login with user "blackhat" and password "password"
    And I register a TOTP secret called "Sec0"
    And I visit "https://auth.test.local:8080/"
    And I login with user "blackhat" and password "password" and I use TOTP token handle "Sec0"
    And I visit "https://auth.test.local:8080/logout?redirect=https://auth.test.local:8080/"
    And I visit "https://auth.test.local:8080/"
    And I set field "username" to "blackhat"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    When I set field "password" to "password"
    And I click on "Sign in"
    Then I get a notification of type "error" with message "Authentication failed. Please double check your credentials."

  @needs-test-config
  Scenario: User is unbanned after a configured amount of time
    Given I visit "https://auth.test.local:8080/"
    And I login with user "blackhat" and password "password"
    And I register a TOTP secret called "Sec0"
    And I visit "https://auth.test.local:8080/"
    And I login with user "blackhat" and password "password" and I use TOTP token handle "Sec0"
    And I visit "https://auth.test.local:8080/logout?redirect=https://auth.test.local:8080/"
    And I visit "https://auth.test.local:8080/"
    And I set field "username" to "blackhat"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please double check your credentials."
    When I wait 6 seconds 
    And I set field "password" to "password"
    And I click on "Sign in"
    And I use "Sec0" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |