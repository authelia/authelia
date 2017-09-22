Feature: Authelia regulates authentication to avoid brute force

  @needs-test-config
  @need-registered-user-blackhat
  Scenario: Attacker tries too many authentication in a short period of time and get banned
    Given I visit "https://auth.test.local:8080/"
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
  @need-registered-user-blackhat
  Scenario: User is unbanned after a configured amount of time
    Given I visit "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"
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
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |